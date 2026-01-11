package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ovh/go-ovh/ovh"
)

// Configuration
var (
	endpoint      string
	appKey        string
	appSecret     string
	consumerKey   string
	zoneName      string
	subDomain     string
	checkInterval string
)

type Record struct {
	ID        uint64 `json:"id"`
	Zone      string `json:"zone"`
	SubDomain string `json:"subDomain"`
	FieldType string `json:"fieldType"`
	Target    string `json:"target"`
	TTL       uint64 `json:"ttl"`
}

type Metrics struct {
	LastCheck      time.Time
	LastUpdate     time.Time
	LastIP         string
	UpdatesSuccess int
	UpdatesFailed  int
	ChecksFailed   int
}

var metrics = &Metrics{}

func main() {
	// Improve log format (standardize timestamp and output)
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, using system environment variables")
	}

	endpoint = os.Getenv("OVH_ENDPOINT")
	appKey = os.Getenv("OVH_APP_KEY")
	appSecret = os.Getenv("OVH_APP_SECRET")
	consumerKey = os.Getenv("OVH_CONSUMER_KEY")
	zoneName = os.Getenv("DNS_ZONE")
	subDomain = os.Getenv("DNS_SUBDOMAIN")
	checkInterval = os.Getenv("CHECK_INTERVAL")

	// Allow subDomain to be empty (for root domain updates)
	if appKey == "" || appSecret == "" || consumerKey == "" || zoneName == "" {
		log.Fatal("[FATAL] Please define environment variables: OVH_ENDPOINT, OVH_APP_KEY, OVH_APP_SECRET, OVH_CONSUMER_KEY, and DNS_ZONE.")
	}

	if endpoint == "" {
		endpoint = "ovh-eu"
	} else {
		validEndpoints := map[string]bool{
			"ovh-eu": true,
			"ovh-us": true,
			"ovh-ca": true,
		}

		if !validEndpoints[endpoint] {
			log.Fatalf("[FATAL] Invalid OVH_ENDPOINT value: %s. Valid options are 'ovh-eu', 'ovh-us', 'ovh-ca'.", endpoint)
		}
	}

	// Define interval
	interval := 5 * time.Minute
	if checkInterval != "" {
		if d, err := time.ParseDuration(checkInterval); err == nil {
			interval = d
		} else {
			log.Fatalf("[FATAL] Invalid CHECK_INTERVAL value: %s. Please use a valid duration format (e.g., '10m', '1h').", checkInterval)
		}
	}

	log.Printf("[INFO] Starting OVH DynDNS for %s.%s (zone: %s)", subDomain, zoneName, zoneName)
	log.Printf("[INFO] Check interval set to %v", interval)

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// First execution immediately
	runDynDNS()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Metrics ticker (every hour)
	metricsTicker := time.NewTicker(1 * time.Hour)
	defer metricsTicker.Stop()

	for {
		select {
		case <-stop:
			log.Println("[INFO] Shutdown signal received, exiting.")
			printMetrics()
			return
		case <-ticker.C:
			runDynDNS()
		case <-metricsTicker.C:
			printMetrics()
		}
	}
}

func runDynDNS() {
	// Capture panics to avoid crashing the container on transient network errors
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CRITICAL] Recovered panic: %v", r)
			metrics.ChecksFailed++
		}
	}()

	metrics.LastCheck = time.Now()

	currentIP, err := getPublicIPWithRetry(3)
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve public IP after retries: %v", err)
		metrics.ChecksFailed++
		return
	}

	client, err := ovh.NewClient(endpoint, appKey, appSecret, consumerKey)
	if err != nil {
		log.Printf("[ERROR] OVH client error: %v", err)
		metrics.ChecksFailed++
		return
	}

	var recordIDs []uint64
	urlFiltering := fmt.Sprintf("/domain/zone/%s/record?fieldType=A&subDomain=%s", zoneName, subDomain)
	if err := client.Get(urlFiltering, &recordIDs); err != nil {
		log.Printf("[ERROR] Failed to search for record: %v", err)
		metrics.ChecksFailed++
		return
	}

	displayDomain := zoneName
	if subDomain != "" {
		displayDomain = subDomain + "." + zoneName
	}

	if len(recordIDs) == 0 {
		log.Printf("[WARN] No 'A' record found for %s in zone %s", displayDomain, zoneName)
		log.Printf("[TIP] Please create the 'A' record manually in your OVH Control Panel first. This application updates existing records but does not create new ones to avoid DNS pollution.")
		metrics.ChecksFailed++
		return
	}

	// Warn if multiple records found
	if len(recordIDs) > 1 {
		log.Printf("[WARN] Multiple A records found (%d) for %s - updating only the first one (ID: %d)",
			len(recordIDs), displayDomain, recordIDs[0])
	}

	recordID := recordIDs[0]
	var record Record
	urlRecord := fmt.Sprintf("/domain/zone/%s/record/%d", zoneName, recordID)
	if err := client.Get(urlRecord, &record); err != nil {
		log.Printf("[ERROR] Failed to read record details: %v", err)
		metrics.ChecksFailed++
		return
	}

	if record.Target == currentIP {
		log.Printf("[INFO] No IP change detected (current: %s)", currentIP)
		metrics.LastIP = currentIP
		return
	}

	log.Printf("[INFO] IP change detected: %s → %s", record.Target, currentIP)

	updateData := map[string]interface{}{"target": currentIP}
	if err := client.Put(urlRecord, updateData, nil); err != nil {
		log.Printf("[ERROR] Update failed: %v", err)
		metrics.UpdatesFailed++
		return
	}

	urlRefresh := fmt.Sprintf("/domain/zone/%s/refresh", zoneName)
	if err := client.Post(urlRefresh, nil, nil); err != nil {
		log.Printf("[ERROR] Zone refresh failed: %v", err)
		metrics.UpdatesFailed++
	} else {
		log.Printf("[INFO] Zone %s refreshed successfully with IP %s", zoneName, currentIP)
		metrics.UpdatesSuccess++
		metrics.LastUpdate = time.Now()
		metrics.LastIP = currentIP
	}
}

// getPublicIPWithRetry attempts to fetch public IP with exponential backoff
func getPublicIPWithRetry(maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		ip, err := getPublicIP()
		if err == nil {
			return ip, nil
		}
		lastErr = err

		if i < maxRetries-1 {
			wait := time.Duration(1<<uint(i)) * time.Second // 1s, 2s, 4s...
			log.Printf("[WARN] IP fetch attempt %d/%d failed: %v - retrying in %v",
				i+1, maxRetries, err, wait)
			time.Sleep(wait)
		}
	}
	return "", fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

func getPublicIP() (string, error) {
	providers := []string{
		"https://api.ipify.org",
		"https://ipv4.icanhazip.com",
		"https://ifconfig.me/ip",
	}

	// Custom transport to force IPv4
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp4", addr)
		},
	}

	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	var lastErr error
	for _, url := range providers {
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("[DEBUG] Provider %s failed: %v", url, err)
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Printf("[DEBUG] Failed to read body from %s: %v", url, err)
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[DEBUG] Provider %s returned status: %d", url, resp.StatusCode)
			lastErr = fmt.Errorf("status code %d", resp.StatusCode)
			continue
		}

		ip := strings.TrimSpace(string(body))
		if ip == "" {
			log.Printf("[DEBUG] Provider %s returned empty IP", url)
			lastErr = fmt.Errorf("empty response")
			continue
		}

		return ip, nil
	}

	return "", fmt.Errorf("all providers failed, last error: %v", lastErr)
}

func printMetrics() {
	log.Printf("[METRICS] Last check: %v | Last IP: %s | Last update: %v",
		metrics.LastCheck.Format(time.RFC3339),
		metrics.LastIP,
		metrics.LastUpdate.Format(time.RFC3339))
	log.Printf("[METRICS] Updates: %d successful, %d failed | Checks failed: %d",
		metrics.UpdatesSuccess,
		metrics.UpdatesFailed,
		metrics.ChecksFailed)
}

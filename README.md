<div align="center">

# 🌐 OVH DynDNS Updater

**Production-Ready Dynamic DNS Client for OVH**

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](Dockerfile)

<p align="center">
    An automated, lightweight, and fault-tolerant service to keep your OVH DNS Zone in sync with your public IP.
    <br/>
    Built with <strong>Go</strong> for performance. Designed for <strong>Docker</strong> for simplicity.
</p>

</div>

---


This application runs as a background service to automatically detect Public IP changes and update your OVH DNS Zone "A" records accordingly. It is built with Go for performance and reliability, featuring IP provider redundancy and Docker support out of the box.

## features

- **Automated DNS Updates**: Monitors public IP and updates OVH DNS records when a change is detected.
- **Failover Redundancy**: Queries multiple IP providers (`ipify`, `icanhazip`, `ifconfig.me`) to ensure connectivity is always accurately verified.
- **IPv4 Forced**: Ensures only valid IPv4 addresses are used for 'A' records, anticipating network stack preferences.
- **Secure Authentication**: Uses OVH native OAuth2 tokens (App Key/Secret & Consumer Key) instead of risky username/password storage.
- **Subdomain Support**: Works with specific subdomains (e.g., `vpn.example.com`) or root domains.
- **Docker Native**: Includes optimized `Dockerfile` (multi-stage build) and `docker-compose.yml` for easy deployment.
- **Metrics & Logging**: Provides clear logs and hourly metrics about successful updates and check status.

## Prerequisites

Before running the application, you need to generate API credentials from OVH.

1.  Go to the [OVH Create Token](https://api.ovh.com/createToken/) page.
2.  Set the validity to "Unlimited".
3.  Grant the following permissions:
    - `GET /domain/zone/*`
    - `PUT /domain/zone/*`
    - `POST /domain/zone/*`
4.  Copy the `App Key`, `App Secret`, and `Consumer Key`.

## Quick Start (Docker Compose)

The easiest way to run the application is using Docker Compose.

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/mateo08c/ovh-dyndns.git
    cd ovh-dyndns
    ```

2.  **Configure Environment**:
    Copy the example configuration file and fill in your credentials.
    ```bash
    cp .env.example .env
    ```
    Edit `.env` with your preferred editor and set your OVH keys and DNS settings.

3.  **Start the Service**:
    ```bash
    docker-compose up -d
    ```

## Manual Installation

If you prefer to run the binary directly on your host machine.

### Build

We provide scripts to easily build for your platform.

**Linux / macOS**
```bash
./scripts/build.sh
```

**Windows (PowerShell)**
```powershell
./scripts/build.ps1
```

The compiled binary will be located in the `build/` directory.

### Run

1.  Ensure your environments variables are set (or a `.env` file is present in the working directory).
2.  Execute the binary:
    ```bash
    ./build/linux_amd64
    ```

## Configuration

The application is configured via environment variables.

| Variable           | Description                                      | Example             |
|--------------------|--------------------------------------------------|---------------------|
| `OVH_ENDPOINT`     | The OVH API endpoint (usually `ovh-eu`).         | `ovh-eu`            |
| `OVH_APP_KEY`      | Your Application Key.                            | `xxxxxxxxxxxx`      |
| `OVH_APP_SECRET`   | Your Application Secret.                         | `xxxxxxxxxxxx`      |
| `OVH_CONSUMER_KEY` | Your Consumer Key.                               | `xxxxxxxxxxxx`      |
| `DNS_ZONE`         | The root domain zone managed in OVH.             | `example.com`       |
| `DNS_SUBDOMAIN`    | The subdomain to update. (e.g., `home`, `vpn`).  | `home`              |
| `CHECK_INTERVAL`   | Time between IP checks (Go duration format).     | `5m` (5 minutes)    |

## Project Structure

```text
.
├── cmd/
│   └── ovh-dyndns/    # Main application entry point
├── scripts/           # Build scripts for Windows and Linux
├── build/             # Output directory for binaries
├── Dockerfile         # Multi-stage Docker build
├── docker-compose.yml # Container orchestration
└── go.mod             # Go dependencies
```

## Contributing

Contributions are welcome! Please feel free to inspect the code, open issues for bugs, or submit pull requests for improvements.

## License

This project is open-source and available under the MIT License.

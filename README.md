# Snipo

A lightweight, self-hosted snippet manager designed for single-user deployments.

> **Note:** This project is intentionally scoped for single-user use. Multi-user features are not planned.

[![CI](https://github.com/MohamedElashri/snipo/actions/workflows/snipo-ci.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/snipo-ci.yml)
[![CI](https://github.com/MohamedElashri/snipo/actions/workflows/snippy-ci.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/snippy-ci.yml)
[![Release](https://github.com/MohamedElashri/snipo/actions/workflows/release.yml/badge.svg)](https://github.com/MohamedElashri/snipo/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![GitHub release](https://img.shields.io/github/v/release/MohamedElashri/snipo?include_prereleases)](https://github.com/MohamedElashri/snipo/releases)

<p align="center">
  <img src="docs/demo.png" alt="Snipo Demo" width="800">
</p>

## Features

- **Fast & Lightweight**: Built with Go and pure SQLite.
- **Powerful Search**: Fuzzy search across titles, descriptions, and file content.
- **Organization**: Organize snippets with folders and tags.
- **Public Sharing**: Share snippets publicly with granular file-level access.
- **Trash & Recovery**: Soft delete mechanism with restore capabilities.
- **GitHub Gist Sync**: Two-way synchronization with GitHub Gists for backup.
- **Developer Friendly**: RESTful API with tokens and granular permissions.
- **Docker Ready**: Easy deployment with Docker and standard or hardened images.
- **Customizable**: Extensive theming support via custom CSS.

## Quick Start
### Docker (Recommended)

```bash
# Create environment file
cat > .env << EOF
SNIPO_MASTER_PASSWORD=your-secure-password
SNIPO_SESSION_SECRET=$(openssl rand -hex 32)
SNIPO_ENCRYPTION_SALT=$(openssl rand -base64 32)
EOF

# Run with Docker Compose
docker compose up -d
```

Or using Docker directly:

```bash
docker run -d \
  -p 8080:8080 \
  -v snipo-data:/data \
  -e SNIPO_MASTER_PASSWORD=your-secure-password \
  -e SNIPO_SESSION_SECRET=$(openssl rand -hex 32) \
  -e SNIPO_ENCRYPTION_SALT=$(openssl rand -base64 32) \
  --name snipo \
  ghcr.io/mohamedelashri/snipo:latest
```

Access at http://localhost:8080

### Binary

```bash
# Download latest release
curl -LO https://github.com/MohamedElashri/snipo/releases/latest/download/snipo_linux_amd64.tar.gz
tar xzf snipo_linux_amd64.tar.gz

# Configure and run
export SNIPO_MASTER_PASSWORD="your-secure-password"
export SNIPO_SESSION_SECRET=$(openssl rand -hex 32)
export SNIPO_ENCRYPTION_SALT=$(openssl rand -base64 32)
./snipo serve
```

## Documentation

- **[Deployment & Configuration](docs/deployment.md)**: Installation, environment variables, reverse proxy setup, and database/API configuration.
- **[Features](docs/features.md)**: In-depth guide to Search, Public Snippets, Gist Sync, and API.
- **[Security](SECURITY.md)**: Security model, authentication modes, and best practices.
- **[Customization](docs/customization.md)**: Theming and custom CSS guide.
- **[Development](docs/Development.md)**: Build instructions and contribution guidelines.
- **[API Spec](docs/openapi.yaml)**: OpenAPI specification.

## License


This project is licensed under the [AGPLv3](LICENSE). Use at your own risk.

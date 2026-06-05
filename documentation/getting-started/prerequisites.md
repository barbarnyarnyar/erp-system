# Prerequisites

System requirements and required software for running the ERP system.

## System Requirements

**Minimum Development Environment:**
- **Operating System**: macOS 10.15+, Ubuntu 20.04+, or Windows 10 with WSL2
- **RAM**: 8GB minimum, 16GB recommended
- **Storage**: 20GB free space for development environment
- **CPU**: 4 cores recommended for optimal Docker performance

**Production Environment:**
- **Kubernetes**: 1.25 or later
- **Cloud Provider**: AWS, GCP, Azure, or on-premises
- **Load Balancer**: With SSL/TLS termination
- **Monitoring**: Prometheus and Grafana compatible

## Required Software

### Git (Version Control)
```bash
# macOS
brew install git

# Ubuntu/Debian
sudo apt-get install git

# Windows (use Git for Windows)
# Download from: https://git-scm.com/download/win

# Verify installation
git --version
```

### Docker and Docker Compose
```bash
# macOS
brew install --cask docker
# Start Docker Desktop application

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io docker-compose-plugin

# Enable Docker service
sudo systemctl enable docker
sudo systemctl start docker

# Add user to docker group (requires logout/login)
sudo usermod -aG docker $USER

# Verify installation
docker --version
docker-compose --version
```

### Go Programming Language
```bash
# macOS
brew install go@1.21

# Ubuntu/Debian
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Add to PATH in ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Verify installation
go version
```

### Node.js and npm
```bash
# macOS
brew install node@18

# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Verify installation
node --version
npm --version
```

### Make (Build Automation)
```bash
# macOS (usually pre-installed)
xcode-select --install

# Ubuntu/Debian
sudo apt-get install build-essential

# Verify installation
make --version
```

## Optional Development Tools

These tools enhance the development experience but aren't required:

### Go-specific Tools
```bash
# Go linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Hot reload for Go development
go install github.com/cosmtrek/air@latest

# Database migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Database Tools
```bash
# PostgreSQL client (optional, for direct database access)
# macOS
brew install postgresql@15

# Ubuntu/Debian
sudo apt-get install postgresql-client-15
```

## Verification

Once all prerequisites are installed, verify your setup:

```bash
# Check versions
git --version
docker --version
docker-compose --version
go version
node --version
make --version

# Test Docker
docker run hello-world
```

If all commands execute successfully, you're ready for [installation](installation.md).
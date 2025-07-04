# Git Stats Golang

[![Go Version](https://img.shields.io/badge/Go-1.21.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com/)
[![Redis](https://img.shields.io/badge/Redis-Cache-red.svg)](https://redis.io/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-orange.svg)](https://prometheus.io/)
[![Grafana](https://img.shields.io/badge/Grafana-Dashboard-yellow.svg)](https://grafana.com/)

> A comprehensive Git statistics application built with Go that provides detailed analytics for GitHub and GitLab repositories with real-time monitoring and caching capabilities.

[🇹🇷 Türkçe README](README-TR.md)

## 🚀 Features

### Core Functionality
- **Multi-Platform Support**: Works with both GitHub and GitLab APIs
- **Repository Analytics**: Comprehensive repository statistics and metrics
- **Commit Analysis**: Detailed commit history and contributor insights
- **Lines of Code Counting**: Accurate LOC calculation for repositories
- **Contributor Statistics**: Detailed contributor analysis and rankings

### Technical Features
- **High-Performance Caching**: Redis-based caching for improved response times
- **Real-time Monitoring**: Prometheus metrics integration
- **Beautiful Dashboards**: Pre-configured Grafana dashboards
- **RESTful API**: Clean and well-documented API endpoints
- **Web Interface**: Modern, responsive web UI
- **CLI Support**: Command-line interface for automation
- **Docker Support**: Fully containerized deployment
- **HTTPS Support**: SSL/TLS encryption for secure communication

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   Go Backend    │    │   Git Providers │
│   (Nginx)       │◄──►│   (API Server)  │◄──►│ GitHub/GitLab   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐              │
         │              │     Redis       │              │
         │              │    (Cache)      │              │
         │              └─────────────────┘              │
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐              │
         └─────────────►│   Prometheus    │◄─────────────┘
                        │   (Metrics)     │
                        └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │    Grafana      │
                        │  (Dashboard)    │
                        └─────────────────┘
```

## 📋 Prerequisites

- **Go 1.21.4+**
- **Docker & Docker Compose**
- **Git**
- **GitHub Personal Access Token** (for GitHub integration)
- **GitLab Personal Access Token** (for GitLab integration)

## 🛠️ Installation

### Option 1: Docker Compose (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/ahmetk3436/git-stats-golang.git
   cd git-stats-golang
   ```

2. **Set up environment variables**:
   ```bash
   export GITHUB_TOKEN="your_github_token_here"
   export GITLAB_TOKEN="your_gitlab_token_here"
   export GITLAB_HOST="https://gitlab.com"  # or your GitLab instance
   export REDIS_PASSWORD="toor"
   ```

3. **Start the application**:
   ```bash
   docker-compose up -d
   ```

4. **Access the services**:
   - **Web Interface**: http://localhost
   - **API**: http://localhost:1323
   - **Prometheus**: http://localhost:9090
   - **Grafana**: http://localhost:3000 (admin/admin)

### Option 2: Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Start Redis** (required for caching):
   ```bash
   docker run -d --name redis -p 6379:6379 redis:latest redis-server --requirepass toor
   ```

3. **Run the application**:
   ```bash
   # API Mode
   go run cmd/main.go api
   
   # CLI Mode
   go run cmd/main.go cli --help
   ```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GITHUB_TOKEN` | GitHub Personal Access Token | - | For GitHub features |
| `GITLAB_TOKEN` | GitLab Personal Access Token | - | For GitLab features |
| `GITLAB_HOST` | GitLab instance URL | `https://gitlab.com` | No |
| `REDIS_HOST` | Redis server address | `redis:6379` | No |
| `REDIS_PASSWORD` | Redis password | `toor` | No |
| `CORS_ALLOWED_ORIGIN` | CORS allowed origins | `*` | No |

### Generating Access Tokens

#### GitHub Token
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token with these scopes:
   - `repo` (for private repositories)
   - `public_repo` (for public repositories)
   - `read:user` (for user information)

#### GitLab Token
1. Go to GitLab Settings → Access Tokens
2. Create token with these scopes:
   - `read_api`
   - `read_repository`
   - `read_user`

## 📚 API Documentation

### GitHub Endpoints

| Method | Endpoint | Description | Parameters |
|--------|----------|-------------|------------|
| GET | `/api/github/repos` | Get all repositories | `owner` (optional) |
| GET | `/api/github/repo` | Get specific repository | `projectID` (required) |
| GET | `/api/github/commits` | Get repository commits | `projectOwner`, `repoName` |
| GET | `/api/github/contributors` | Get repository contributors | `projectOwner`, `repoName` |
| GET | `/api/github/loc` | Get lines of code | `projectOwner`, `repoName` |

### GitLab Endpoints

| Method | Endpoint | Description | Parameters |
|--------|----------|-------------|------------|
| GET | `/api/gitlab/repos` | Get all repositories | `owner` (optional) |
| GET | `/api/gitlab/repo` | Get specific repository | `projectID` (required) |
| GET | `/api/gitlab/commits` | Get repository commits | `projectOwner`, `repoName` |

### Example API Calls

```bash
# Get all GitHub repositories
curl "http://localhost:1323/api/github/repos"

# Get specific repository
curl "http://localhost:1323/api/github/repo?projectID=owner/repo-name"

# Get repository commits
curl "http://localhost:1323/api/github/commits?projectOwner=owner&repoName=repo-name"

# Get repository contributors
curl "http://localhost:1323/api/github/contributors?projectOwner=owner&repoName=repo-name"
```

## 🖥️ CLI Usage

```bash
# GitHub operations
go run cmd/main.go cli --github-token="your_token" --help

# GitLab operations
go run cmd/main.go cli --gitlab-token="your_token" --gitlab-host="https://gitlab.com" --help

# Get repository information
go run cmd/main.go cli --github-token="your_token" repo --owner="username" --repo="repository"
```

## 📊 Monitoring & Metrics

### Prometheus Metrics

The application exposes the following metrics at `/metrics`:

- `gits_api_calls_total`: Total number of API calls
- `gits_repository_fetches_total`: Total repository fetch attempts
- `gits_api_call_duration_seconds`: API call duration histogram

### Grafana Dashboard

Pre-configured dashboard includes:
- API response times
- Request rates
- Error rates
- Cache hit ratios
- Repository statistics

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/api/
go test ./pkg/repository/
```

## 🏗️ Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── main.go            # Main application
│   ├── cert.pem           # SSL certificate (dev only)
│   └── key.pem            # SSL private key (dev only)
├── pkg/                    # Public packages
│   ├── api/               # HTTP API handlers
│   ├── cli/               # CLI commands
│   ├── common_types/      # Shared data structures
│   ├── interfaces/        # Interface definitions
│   ├── prometheus/        # Metrics definitions
│   └── repository/        # Git provider implementations
├── internal/              # Private packages
│   └── inmemory_db.go     # Redis client
├── web/                   # Frontend assets
│   ├── index.html         # Web interface
│   └── api.js             # Frontend JavaScript
├── yaml/                  # Configuration files
│   ├── prometheus.yml     # Prometheus config
│   ├── dashboard.yml      # Grafana dashboard config
│   └── golang.json        # Grafana dashboard JSON
├── docker-compose.yaml    # Docker composition
├── Dockerfile             # Application container
└── Dockerfile-Redis       # Redis container
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit your changes**: `git commit -m 'Add amazing feature'`
4. **Push to the branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

### Development Guidelines

- Follow Go best practices and conventions
- Write comprehensive tests for new features
- Update documentation for API changes
- Use structured logging with logrus
- Implement proper error handling
- Add Prometheus metrics for new endpoints

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Go](https://golang.org/) - The programming language
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router
- [Redis](https://redis.io/) - In-memory data structure store
- [Prometheus](https://prometheus.io/) - Monitoring system
- [Grafana](https://grafana.com/) - Analytics platform
- [Docker](https://www.docker.com/) - Containerization platform

## 📞 Support

If you have any questions or need help, please:

1. Check the [documentation](#-api-documentation)
2. Search existing [issues](https://github.com/ahmetk3436/git-stats-golang/issues)
3. Create a new [issue](https://github.com/ahmetk3436/git-stats-golang/issues/new)

---

**Made with ❤️ by [Ahmet Coşkun Kızılkaya](https://github.com/ahmetk3436)**
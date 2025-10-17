# Uptime Application

A modern uptime monitoring application built with Go, Next.js, and Kubernetes.

## Architecture

The application consists of two main services:

- **Web Frontend** (Next.js) - User interface running on port 3000
- **API Services** (Go) - Backend API services running on port 8082

The web frontend connects directly to the API services for a simplified, efficient architecture.

## Quick Start

### Prerequisites

- Docker Desktop with Kubernetes enabled
- Tilt CLI
- kubectl configured for your local cluster

### Development

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd uptime-application
   ```

2. Start the development environment:
   ```bash
   tilt up --port=8090
   ```

3. Access the application:
   - Web Frontend: http://localhost:3000
   - API Services: http://localhost:8082
   - Tilt Dashboard: http://localhost:8090

## Services

### Web Frontend
- **Technology**: Next.js 14 with TypeScript
- **Port**: 3000
- **Features**: Modern React-based UI with server-side rendering

### API Services
- **Technology**: Go with Gin framework
- **Port**: 8082
- **Features**: RESTful API for uptime monitoring, user management, and analytics

## Development Workflow

- Code changes trigger automatic rebuilds and restarts
- Web frontend runs with hot reloading
- Logs are available in the Tilt UI
- Health checks ensure services are running properly

## Infrastructure

See the [Infrastructure README](./infra/README.md) for detailed deployment instructions and configuration options.

## Contributing

1. Make changes in your local development environment
2. Test using the Tilt development setup
3. Ensure all services pass health checks
4. Submit a pull request

## License

[Add your license information here]
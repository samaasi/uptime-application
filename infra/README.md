# Uptime Application Infrastructure

This directory contains the Kubernetes infrastructure and Docker configurations for the Uptime Application.

## Directory Structure

```
infra/
├── development/
│   ├── docker/
│   │   ├── api-service.Dockerfile
│   │   └── web.Dockerfile
│   └── k8s/
│       ├── app-config.yaml
│       ├── api-service-deployment.yaml
│       └── web-deployment.yaml
├── production/
│   └── k8s/
│       ├── app-config.yaml
│       ├── api-services-deployment.yaml
│       ├── web-deployment.yaml
│       └── ingress.yaml
└── README.md
```

## Prerequisites

### Development
- Docker Desktop with Kubernetes enabled
- Tilt CLI (`curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash`)
- kubectl configured for your local cluster
- External services (PostgreSQL, Redis, ClickHouse) running and accessible

### Production
- Kubernetes cluster (EKS, GKE, AKS, or self-managed)
- kubectl configured for your production cluster
- Ingress controller (nginx-ingress recommended)
- Cert-manager for SSL certificates
- External services (PostgreSQL, Redis, ClickHouse) configured

## Development Setup

### 1. Start External Services

Ensure you have PostgreSQL, Redis, and ClickHouse running and accessible. Update the connection details in `development/k8s/app-config.yaml`.

### 2. Configure Environment

Update the configuration in `infra/development/k8s/app-config.yaml`:
- Database connection details
- Redis connection details
- ClickHouse connection details
- API URLs and ports

### 3. Start Development Environment

```bash
# From the project root
tilt up
```

This will:
- Build Docker images for all services
- Deploy to your local Kubernetes cluster
- Set up port forwarding
- Enable live reloading for development

### 4. Access Services

- Web Frontend: http://localhost:3000
- API Services: http://localhost:8082

### 5. Development Workflow

- Code changes in Go services trigger automatic rebuilds and restarts
- Web frontend runs in development mode with hot reloading
- Logs are available in the Tilt UI

## Production Deployment

### 1. Prepare Environment

1. **Create namespace:**
   ```bash
   kubectl apply -f infra/production/k8s/app-config.yaml
   ```

2. **Update configuration:**
   - Edit `infra/production/k8s/app-config.yaml`
   - Update all `CHANGE_ME_*` values in the secrets
   - Configure your domain names in the ingress

3. **Build and push images:**
   ```bash
   # Build production images
   docker build -f infra/development/docker/api-service.Dockerfile -t uptime/api-services:latest .
   docker build -f infra/development/docker/web.Dockerfile -t uptime/web:latest .
   
   # Push to your registry
   docker tag uptime/api-services:latest your-registry/uptime/api-services:latest
   docker tag uptime/web:latest your-registry/uptime/web:latest
   
   docker push your-registry/uptime/api-services:latest
   docker push your-registry/uptime/web:latest
   ```

### 2. Deploy Services

```bash
# Apply all production manifests
kubectl apply -f infra/production/k8s/

# Check deployment status
kubectl get pods -n uptime-prod
kubectl get services -n uptime-prod
kubectl get ingress -n uptime-prod
```

### 3. Configure DNS

Point your domain names to the ingress controller's external IP:
- yourdomain.com → Ingress External IP
- www.yourdomain.com → Ingress External IP  
- api.yourdomain.com → Ingress External IP (routes directly to api-services)

### 4. SSL Certificates

If using cert-manager, certificates will be automatically provisioned. Otherwise, manually configure SSL certificates.

## Configuration

### Environment Variables

#### Database
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_NAME`: Database name
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_SSL_MODE`: SSL mode (disable for dev, require for prod)

#### Redis
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port (default: 6379)
- `REDIS_DB`: Redis database number
- `REDIS_PASSWORD`: Redis password (if required)

#### ClickHouse
- `CLICKHOUSE_HOST`: ClickHouse host
- `CLICKHOUSE_PORT`: ClickHouse port (default: 9000)
- `CLICKHOUSE_DATABASE`: ClickHouse database name
- `CLICKHOUSE_USER`: ClickHouse username
- `CLICKHOUSE_PASSWORD`: ClickHouse password

#### Application
- `APP_ENV`: Environment (development/production)
- `APP_DEBUG`: Debug mode (true/false)
- `LOG_LEVEL`: Logging level (debug/info/warn/error)
- `JWT_SECRET`: JWT signing secret (must be strong in production)
- `ENCRYPTION_KEY`: Encryption key (exactly 32 characters)

## Monitoring and Observability

### Health Checks

All services expose health check endpoints:
- API Services: `/health` and `/ready`
- Web: `/api/health`

### Metrics

Services are configured with Prometheus annotations for metrics collection:
- API Services: `:8082/metrics`
- Web: `:3000/api/metrics`

### Logs

Use kubectl to view logs:
```bash
# View logs for a specific service
kubectl logs -f deployment/api-services -n uptime-prod

# View logs for all pods with a label
kubectl logs -f -l app=api-services -n uptime-prod
```

## Scaling

### Development
Development environment runs single replicas for easier debugging.

### Production
Production environment includes:
- Horizontal Pod Autoscalers (HPA) for automatic scaling
- Pod Disruption Budgets (PDB) for high availability
- Anti-affinity rules for pod distribution

Manual scaling:
```bash
kubectl scale deployment api-services --replicas=5 -n uptime-prod
```

## Security

### Production Security Features
- Non-root containers
- Read-only root filesystems
- Security contexts with dropped capabilities
- Network policies for traffic isolation
- Resource limits and requests
- Pod security standards compliance

### Secrets Management
- Kubernetes secrets for sensitive data
- Environment-specific configurations
- Separate secrets for each environment

## Troubleshooting

### Common Issues

1. **Pods not starting:**
   ```bash
   kubectl describe pod <pod-name> -n uptime-prod
   kubectl logs <pod-name> -n uptime-prod
   ```

2. **Service connectivity issues:**
   ```bash
   kubectl get endpoints -n uptime-prod
   kubectl exec -it <pod-name> -n uptime-prod -- nslookup <service-name>
   ```

3. **Ingress not working:**
   ```bash
   kubectl describe ingress uptime-ingress -n uptime-prod
   kubectl get events -n uptime-prod
   ```

### Debug Commands

```bash
# Check cluster status
kubectl cluster-info

# Check node resources
kubectl top nodes

# Check pod resources
kubectl top pods -n uptime-prod

# Get all resources in namespace
kubectl get all -n uptime-prod

# Check events
kubectl get events -n uptime-prod --sort-by='.lastTimestamp'
```

## Backup and Recovery

### Database Backups
Configure regular backups for your external PostgreSQL and ClickHouse databases.

### Configuration Backups
Keep your Kubernetes manifests in version control and backup your cluster configuration.

### Disaster Recovery
Document your disaster recovery procedures including:
- Database restoration
- Cluster recreation
- DNS failover
- Certificate renewal

## Contributing

When making changes to the infrastructure:
1. Test changes in development environment first
2. Update documentation
3. Follow security best practices
4. Test scaling and failover scenarios
5. Update monitoring and alerting as needed
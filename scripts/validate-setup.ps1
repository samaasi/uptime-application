# Uptime Application Development Setup Validation Script

Write-Host "=== Uptime Application Development Setup Validation ===" -ForegroundColor Green
Write-Host ""

# Check if Docker is running
Write-Host "Checking Docker..." -ForegroundColor Yellow
try {
    $dockerVersion = docker version --format "{{.Server.Version}}" 2>$null
    if ($dockerVersion) {
        Write-Host "✓ Docker is running (version: $dockerVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Docker is not running" -ForegroundColor Red
        Write-Host "  Please start Docker Desktop" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Docker is not installed or not in PATH" -ForegroundColor Red
}

# Check if Kubernetes is enabled in Docker Desktop
Write-Host "Checking Kubernetes..." -ForegroundColor Yellow
try {
    $kubectlVersion = kubectl version --client --short 2>$null
    if ($kubectlVersion) {
        Write-Host "✓ kubectl is installed" -ForegroundColor Green
        
        # Check if cluster is accessible
        $clusterInfo = kubectl cluster-info 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Kubernetes cluster is accessible" -ForegroundColor Green
        } else {
            Write-Host "✗ Kubernetes cluster is not accessible" -ForegroundColor Red
            Write-Host "  Please enable Kubernetes in Docker Desktop:" -ForegroundColor Yellow
            Write-Host "  1. Open Docker Desktop" -ForegroundColor Yellow
            Write-Host "  2. Go to Settings > Kubernetes" -ForegroundColor Yellow
            Write-Host "  3. Check 'Enable Kubernetes'" -ForegroundColor Yellow
            Write-Host "  4. Click 'Apply & Restart'" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ kubectl is not installed or not in PATH" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ kubectl is not installed or not in PATH" -ForegroundColor Red
}

# Check if Tilt is installed
Write-Host "Checking Tilt..." -ForegroundColor Yellow
try {
    $tiltVersion = tilt version 2>$null
    if ($tiltVersion) {
        Write-Host "✓ Tilt is installed ($tiltVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Tilt is not installed" -ForegroundColor Red
        Write-Host "  Install Tilt using:" -ForegroundColor Yellow
        Write-Host "  iex ((new-object net.webclient).downloadstring('https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.ps1'))" -ForegroundColor Cyan
    }
} catch {
    Write-Host "✗ Tilt is not installed or not in PATH" -ForegroundColor Red
}

# Check if Go is installed (for building services)
Write-Host "Checking Go..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>$null
    if ($goVersion) {
        Write-Host "✓ Go is installed ($goVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Go is not installed" -ForegroundColor Red
        Write-Host "  Please install Go from https://golang.org/dl/" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Go is not installed or not in PATH" -ForegroundColor Red
}

# Check if Node.js is installed (for web frontend)
Write-Host "Checking Node.js..." -ForegroundColor Yellow
try {
    $nodeVersion = node --version 2>$null
    if ($nodeVersion) {
        Write-Host "✓ Node.js is installed ($nodeVersion)" -ForegroundColor Green
        
        # Check npm
        $npmVersion = npm --version 2>$null
        if ($npmVersion) {
            Write-Host "✓ npm is installed ($npmVersion)" -ForegroundColor Green
        }
    } else {
        Write-Host "✗ Node.js is not installed" -ForegroundColor Red
        Write-Host "  Please install Node.js from https://nodejs.org/" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Node.js is not installed or not in PATH" -ForegroundColor Red
}

# Check if required directories exist
Write-Host "Checking project structure..." -ForegroundColor Yellow
$requiredDirs = @(
    "services/api-gateway",
    "services/api-services", 
    "web",
    "infra/development/docker",
    "infra/development/k8s",
    "infra/production/k8s"
)

foreach ($dir in $requiredDirs) {
    if (Test-Path $dir) {
        Write-Host "✓ Directory exists: $dir" -ForegroundColor Green
    } else {
        Write-Host "✗ Directory missing: $dir" -ForegroundColor Red
    }
}

# Check if required files exist
Write-Host "Checking required files..." -ForegroundColor Yellow
$requiredFiles = @(
    "Tiltfile",
    "infra/development/docker/api-gateway.Dockerfile",
    "infra/development/docker/api-service.Dockerfile", 
    "infra/development/docker/web.Dockerfile",
    "infra/development/k8s/app-config.yaml",
    "infra/development/k8s/api-gateway-deployment.yaml",
    "infra/development/k8s/api-service-deployment.yaml",
    "infra/development/k8s/web-deployment.yaml"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✓ File exists: $file" -ForegroundColor Green
    } else {
        Write-Host "✗ File missing: $file" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Next Steps ===" -ForegroundColor Green
Write-Host ""
Write-Host "If all checks pass, you can start the development environment with:" -ForegroundColor Yellow
Write-Host "  tilt up" -ForegroundColor Cyan
Write-Host ""
Write-Host "This will:" -ForegroundColor Yellow
Write-Host "  • Build Docker images for all services" -ForegroundColor White
Write-Host "  • Deploy to your local Kubernetes cluster" -ForegroundColor White
Write-Host "  • Set up port forwarding" -ForegroundColor White
Write-Host "  • Enable live reloading for development" -ForegroundColor White
Write-Host ""
Write-Host "Access your services at:" -ForegroundColor Yellow
Write-Host "  • Web Frontend: http://localhost:3000" -ForegroundColor Cyan
Write-Host "  • API Gateway: http://localhost:8081" -ForegroundColor Cyan
Write-Host "  • API Services: http://localhost:8082" -ForegroundColor Cyan
Write-Host ""
Write-Host "Before starting, make sure you have configured:" -ForegroundColor Yellow
Write-Host "  • PostgreSQL connection in infra/development/k8s/app-config.yaml" -ForegroundColor White
Write-Host "  • Redis connection in infra/development/k8s/app-config.yaml" -ForegroundColor White
Write-Host "  • ClickHouse connection in infra/development/k8s/app-config.yaml" -ForegroundColor White
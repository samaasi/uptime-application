Write-Host "=== Uptime Application Development Setup Validation ===" -ForegroundColor Green
Write-Host ""

Write-Host "Checking Docker..." -ForegroundColor Yellow
try {
    $dockerVersion = docker version --format "{{.Server.Version}}" 2>$null
    if ($dockerVersion) {
        Write-Host "✓ Docker is running (version: $dockerVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Docker is not running" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Docker is not installed" -ForegroundColor Red
}

Write-Host "Checking Kubernetes..." -ForegroundColor Yellow
try {
    $kubectlVersion = kubectl version --client --short 2>$null
    if ($kubectlVersion) {
        Write-Host "✓ kubectl is installed" -ForegroundColor Green
        $clusterInfo = kubectl cluster-info 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Kubernetes cluster is accessible" -ForegroundColor Green
        } else {
            Write-Host "✗ Kubernetes cluster is not accessible" -ForegroundColor Red
            Write-Host "  Please enable Kubernetes in Docker Desktop" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ kubectl is not installed" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ kubectl is not installed" -ForegroundColor Red
}

Write-Host "Checking Tilt..." -ForegroundColor Yellow
try {
    $tiltVersion = tilt version 2>$null
    if ($tiltVersion) {
        Write-Host "✓ Tilt is installed ($tiltVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Tilt is not installed" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Tilt is not installed" -ForegroundColor Red
}

Write-Host "Checking Go..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>$null
    if ($goVersion) {
        Write-Host "✓ Go is installed ($goVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Go is not installed" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Go is not installed" -ForegroundColor Red
}

Write-Host "Checking Node.js..." -ForegroundColor Yellow
try {
    $nodeVersion = node --version 2>$null
    if ($nodeVersion) {
        Write-Host "✓ Node.js is installed ($nodeVersion)" -ForegroundColor Green
    } else {
        Write-Host "✗ Node.js is not installed" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Node.js is not installed" -ForegroundColor Red
}

Write-Host "Checking project structure..." -ForegroundColor Yellow
$requiredDirs = @("services/api-gateway", "services/api-services", "web", "infra/development/docker", "infra/development/k8s", "infra/production/k8s")

foreach ($dir in $requiredDirs) {
    if (Test-Path $dir) {
        Write-Host "✓ Directory exists: $dir" -ForegroundColor Green
    } else {
        Write-Host "✗ Directory missing: $dir" -ForegroundColor Red
    }
}

Write-Host "Checking required files..." -ForegroundColor Yellow
$requiredFiles = @("Tiltfile", "infra/development/docker/api-gateway.Dockerfile", "infra/development/docker/api-service.Dockerfile", "infra/development/docker/web.Dockerfile", "infra/development/k8s/app-config.yaml")

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✓ File exists: $file" -ForegroundColor Green
    } else {
        Write-Host "✗ File missing: $file" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Next Steps ===" -ForegroundColor Green
Write-Host "If all checks pass, start the development environment with: tilt up" -ForegroundColor Yellow
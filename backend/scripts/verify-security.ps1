# PowerShell script to verify Docker image security

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Docker Security Verification" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

$IMAGE_NAME = "file-sharing-backend:latest"

# Build the image
Write-Host "📦 Building production image..." -ForegroundColor Yellow
docker build -t $IMAGE_NAME . 2>&1 | Out-Null

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

# Check if Trivy is installed
$trivyInstalled = Get-Command trivy -ErrorAction SilentlyContinue
if (-not $trivyInstalled) {
    Write-Host "⚠️  Trivy not installed." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Installing Trivy via Chocolatey..." -ForegroundColor Yellow
    
    $chocoInstalled = Get-Command choco -ErrorAction SilentlyContinue
    if ($chocoInstalled) {
        choco install trivy -y
    } else {
        Write-Host "❌ Chocolatey not installed. Please install manually:" -ForegroundColor Red
        Write-Host "  1. Install Chocolatey: https://chocolatey.org/install"
        Write-Host "  2. Run: choco install trivy"
        Write-Host ""
        Write-Host "Or download from: https://github.com/aquasecurity/trivy/releases"
        exit 1
    }
}

# Scan the image
Write-Host ""
Write-Host "🔍 Scanning image with Trivy..." -ForegroundColor Yellow
Write-Host ""

trivy image --severity HIGH,CRITICAL $IMAGE_NAME

# Check exit code
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "✅ PASS: No HIGH/CRITICAL vulnerabilities found!" -ForegroundColor Green
    Write-Host "Your production image is secure." -ForegroundColor Green
    Write-Host "======================================" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "❌ FAIL: Vulnerabilities found" -ForegroundColor Red
    Write-Host "Please review and update base images." -ForegroundColor Red
    Write-Host "======================================" -ForegroundColor Cyan
    exit 1
}

# Show image size
Write-Host ""
Write-Host "📊 Image Statistics:" -ForegroundColor Cyan
docker images $IMAGE_NAME --format "Size: {{.Size}}"

# Show image layers (first 10)
Write-Host ""
Write-Host "📋 Image Layers (top 10):" -ForegroundColor Cyan
docker history $IMAGE_NAME --no-trunc --format "{{.CreatedBy}}" | Select-Object -First 10

Write-Host ""
Write-Host "✅ Security verification complete!" -ForegroundColor Green

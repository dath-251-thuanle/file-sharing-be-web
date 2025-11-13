#!/bin/bash
# Script to verify Docker image security

set -e

echo "======================================"
echo "Docker Security Verification"
echo "======================================"
echo ""

IMAGE_NAME="file-sharing-backend:latest"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build the image
echo "📦 Building production image..."
docker build -t $IMAGE_NAME .

# Check if Trivy is installed
if ! command -v trivy &> /dev/null; then
    echo -e "${YELLOW}⚠️  Trivy not installed. Installing...${NC}"
    # Windows: choco install trivy
    # Linux: wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
    # Mac: brew install trivy
    echo "Please install Trivy manually:"
    echo "  Windows: choco install trivy"
    echo "  Linux: See https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
    echo "  Mac: brew install trivy"
    exit 1
fi

# Scan the image
echo ""
echo "🔍 Scanning image with Trivy..."
echo ""

trivy image --severity HIGH,CRITICAL $IMAGE_NAME

# Get vulnerability count
VULN_COUNT=$(trivy image --severity HIGH,CRITICAL --format json $IMAGE_NAME 2>/dev/null | grep -o '"VulnerabilityID"' | wc -l)

echo ""
echo "======================================"
if [ "$VULN_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS: No HIGH/CRITICAL vulnerabilities found!${NC}"
    echo "Your production image is secure."
else
    echo -e "${RED}❌ FAIL: Found $VULN_COUNT HIGH/CRITICAL vulnerabilities${NC}"
    echo "Please review and update base images."
    exit 1
fi
echo "======================================"

# Show image size
echo ""
echo "📊 Image Statistics:"
docker images $IMAGE_NAME --format "Size: {{.Size}}"

# Show image layers
echo ""
echo "📋 Image Layers:"
docker history $IMAGE_NAME --no-trunc --format "{{.CreatedBy}}" | head -10

echo ""
echo "✅ Security verification complete!"

# Docker Security Best Practices

## 🔐 Security Issues Fixed

### Problem: Alpine/Debian base images có vulnerabilities

### Solutions được implement:

## 1️⃣ **Dockerfile** (Production) - Distroless Image ✅ RECOMMENDED

```dockerfile
# Uses gcr.io/distroless/static-debian12:nonroot
# ✅ Minimal attack surface (no shell, no package manager)
# ✅ Non-root user by default (UID 65532)
# ✅ Latest security patches
# ✅ ~2MB final image
```

**Pros:**
- Ít vulnerabilities nhất (không có shell, package manager)
- Image size nhỏ (~2MB vs ~40MB Alpine)
- Google maintains và update thường xuyên
- Non-root by default

**Cons:**
- Không có shell → debug khó hơn
- Không có healthcheck với wget/curl

**Use for:** Production deployments

---

## 2️⃣ **Dockerfile.alpine** - Hardened Alpine ⚠️

```dockerfile
# Uses alpine:3.20 (latest stable)
# ✅ Latest security patches (apk upgrade)
# ✅ Non-root user
# ✅ Healthcheck support
# ⚠️ Có một số vulnerabilities (ít hơn old versions)
```

**Pros:**
- Có shell → dễ debug
- Có healthcheck
- Image size vừa phải (~15MB)
- Familiar cho developers

**Cons:**
- Vẫn có một số CVEs (thường low/medium)
- Cần update thường xuyên

**Use for:** Development hoặc khi cần debug

---

## 3️⃣ **Dockerfile.dev** - Development với Debian

```dockerfile
# Uses golang:1.23-bookworm
# ✅ Có đầy đủ tools
# ✅ Hot reload với Air
# ⚠️ Image size lớn (~800MB)
```

**Use for:** Local development only

---

## 📊 So sánh:

| Image Type | Size | Vulnerabilities | Debug | Best For |
|------------|------|-----------------|-------|----------|
| **Distroless** | ~2MB | ✅ Minimal | ❌ No shell | Production |
| **Alpine (hardened)** | ~15MB | ⚠️ Some | ✅ Yes | Dev/Staging |
| **Debian/Bookworm** | ~800MB | ⚠️ More | ✅ Yes | Development |

---

## 🚀 Cách sử dụng:

### Production (Distroless) - RECOMMENDED

```bash
# Build
docker build -t file-sharing-backend:latest .

# Hoặc dùng secure variant
docker build -f Dockerfile.secure -t file-sharing-backend:secure .

# Run
docker run -d \
  --name app \
  -p 8080:8080 \
  --env-file .env \
  file-sharing-backend:latest
```

### Development (Debian)

```bash
docker-compose --profile dev up -d
```

### Staging (Alpine hardened)

```bash
docker build -f Dockerfile.alpine -t file-sharing-backend:alpine .
docker run -d -p 8080:8080 --env-file .env file-sharing-backend:alpine
```

---

## 🔍 Scan for vulnerabilities:

### Trivy (Recommended)

```bash
# Install Trivy
# Windows (Chocolatey)
choco install trivy

# Scan image
trivy image file-sharing-backend:latest

# Scan với severity filter
trivy image --severity HIGH,CRITICAL file-sharing-backend:latest

# Generate report
trivy image -f json -o report.json file-sharing-backend:latest
```

### Docker Scout

```bash
# Enable Docker Scout
docker scout enroll

# Scan image
docker scout cves file-sharing-backend:latest

# Compare with base image
docker scout compare --to golang:1.23-alpine3.20 file-sharing-backend:latest
```

### Snyk

```bash
# Install Snyk CLI
npm install -g snyk

# Login
snyk auth

# Scan Docker image
snyk container test file-sharing-backend:latest

# Monitor in Snyk dashboard
snyk container monitor file-sharing-backend:latest
```

---

## 🛡️ Additional Security Measures:

### 1. Update docker-compose.yml

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile  # Uses distroless
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE
    read_only: true
    tmpfs:
      - /tmp
    volumes:
      - app_storage:/app/storage/uploads:rw
```

### 2. Add security scanning to CI/CD

Update `.github/workflows/build.yml`:

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'CRITICAL,HIGH'

- name: Upload Trivy results to GitHub Security
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: 'trivy-results.sarif'

- name: Fail on critical vulnerabilities
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
    exit-code: '1'
    severity: 'CRITICAL'
```

### 3. Regular updates

```bash
# Update base images weekly
docker pull golang:1.23-bookworm
docker pull gcr.io/distroless/static-debian12:nonroot
docker pull alpine:3.20

# Rebuild
docker-compose build --no-cache
```

### 4. Use .dockerignore

Already configured to exclude:
- `.git`, `.env`, test files
- Build artifacts, logs
- Reduces attack surface

---

## 📝 Best Practices Checklist:

- ✅ Use distroless for production
- ✅ Multi-stage builds (builder + final)
- ✅ Non-root user
- ✅ Minimal base image
- ✅ No secrets in Dockerfile
- ✅ Regular security scans
- ✅ Update base images regularly
- ✅ Use specific tags (not `latest`)
- ✅ Read-only filesystem when possible
- ✅ Drop unnecessary capabilities
- ✅ Sign images (Docker Content Trust)

---

## 🔄 Migration Guide:

### Current → Distroless (Recommended)

1. Update docker-compose.yml:
```yaml
app:
  build:
    context: .
    dockerfile: Dockerfile  # Already uses distroless
```

2. Remove healthcheck from Dockerfile (distroless không có wget)

3. Add healthcheck in docker-compose.yml:
```yaml
healthcheck:
  test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"]
  # Runs from host, not container
```

### For debugging distroless:

```bash
# Use debug variant (has shell)
FROM gcr.io/distroless/static-debian12:debug-nonroot

# Exec into container
docker run -it --entrypoint sh gcr.io/distroless/static-debian12:debug-nonroot
```

---

## 📚 Resources:

- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [Trivy Scanner](https://github.com/aquasecurity/trivy)
- [OWASP Docker Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)

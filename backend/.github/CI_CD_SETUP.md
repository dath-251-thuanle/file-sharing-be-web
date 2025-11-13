# CI/CD Configuration Guide

## 📋 Overview

CI/CD được thiết lập với GitHub Actions gồm 7 workflows chính:

### 1. **CI (Continuous Integration)** - `ci.yml`

- Trigger: Push/PR vào main, develop, feature/*
- Jobs:
  - ✅ Test với PostgreSQL
  - ✅ Lint (golangci-lint)
  - ✅ Security scan (gosec)
  - ✅ Coverage report

### 2. **Build** - `build.yml`

- Trigger: Push vào main/develop, tags
- Jobs:
  - 🐳 Build Docker image
  - 📦 Push to GitHub Container Registry
  - 🏷️ Multi-platform (amd64, arm64)

### 3. **Deploy Production** - `deploy.yml`

- Trigger: Push vào main, tags v*.*.*
- Jobs:
  - 🚀 Deploy to production server
  - 🔄 Auto rollback on failure
  - 💬 Slack notification

### 4. **Deploy Staging** - `deploy-staging.yml`

- Trigger: Push vào develop
- Jobs:
  - 🧪 Deploy to staging environment
  - ✅ Smoke tests
  - 💬 Team notification

### 5. **Release** - `release.yml`

- Trigger: Git tag v*.*.*
- Jobs:
  - 📦 Build multi-platform binaries
  - 📝 Generate changelog
  - 🎉 Create GitHub Release
  - 🐳 Tag Docker images

### 6. **Database Migration** - `migrate.yml`

- Trigger: Manual (workflow_dispatch)
- Jobs:
  - 🗄️ Apply/rollback migrations
  - ✅ Verify migration

### 7. **Cleanup** - `cleanup.yml`

- Trigger: Cron (daily 2 AM) hoặc manual
- Jobs:
  - 🧹 Cleanup expired files
  - ✅ Health check

---

## 🔧 Setup Instructions

### Step 1: Configure GitHub Secrets

Vào **Settings → Secrets and variables → Actions** và thêm:

#### Production Secrets:

```
DEPLOY_HOST=your-production-server.com
DEPLOY_USER=deploy
DEPLOY_SSH_KEY=<your-private-ssh-key>
DEPLOY_PORT=22

DB_HOST=your-db-host.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<secure-password>
DB_NAME=file_sharing_db

CLEANUP_SECRET=<your-cleanup-secret>
API_URL=https://api.filesharing.com

SLACK_WEBHOOK=<slack-webhook-url>
```

#### Staging Secrets:

```
STAGING_HOST=staging.filesharing.com
STAGING_USER=deploy
STAGING_SSH_KEY=<staging-private-ssh-key>
```

### Step 2: Configure Environments

Vào **Settings → Environments** và tạo:

1. **production**

   - Protection rules:
     - ✅ Required reviewers (1-2 người)
     - ✅ Wait timer: 5 minutes
   - Environment secrets (như trên)
2. **staging**

   - Không cần protection rules
   - Environment secrets cho staging

### Step 3: Enable GitHub Container Registry

```bash
# 1. Generate Personal Access Token
# Settings → Developer settings → Personal access tokens → Generate new token
# Permissions: write:packages, read:packages

# 2. Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# 3. Images sẽ được push tự động vào:
# ghcr.io/yourusername/file-sharing-backend:latest
# ghcr.io/yourusername/file-sharing-backend:v1.0.0
```

### Step 4: Setup Production Server

#### Install Docker & Docker Compose

```bash
# SSH vào server
ssh deploy@your-production-server.com

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### Setup deployment directory

```bash
sudo mkdir -p /opt/file-sharing-backend
sudo chown deploy:deploy /opt/file-sharing-backend
cd /opt/file-sharing-backend

# Create docker-compose.yml
cat > docker-compose.yml <<EOF
version: '3.8'

services:
  app:
    image: ghcr.io/yourusername/file-sharing-backend:latest
    ports:
      - "8080:8080"
    env_file:
      - .env
    volumes:
      - ./storage:/app/storage
    restart: unless-stopped
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: file_sharing_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: \${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
EOF

# Create .env file
cp .env.example .env
# Edit với production values
```

#### Setup SSH Key

```bash
# Trên máy local
ssh-keygen -t ed25519 -C "deploy@production" -f ~/.ssh/deploy_key

# Copy public key lên server
ssh-copy-id -i ~/.ssh/deploy_key.pub deploy@your-production-server.com

# Add private key vào GitHub Secrets (DEPLOY_SSH_KEY)
cat ~/.ssh/deploy_key
```

### Step 5: Setup Nginx (Reverse Proxy)

```bash
# Install Nginx
sudo apt install nginx

# Create config
sudo nano /etc/nginx/sites-available/file-sharing

# Paste:
server {
    listen 80;
    server_name api.filesharing.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Enable site
sudo ln -s /etc/nginx/sites-available/file-sharing /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Setup SSL với Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.filesharing.com
```

---

## 🚀 Usage

### Deploy to Production

#### Option 1: Automatic (via Git tag)

```bash
git tag v1.0.0
git push origin v1.0.0
# → Triggers: build.yml, release.yml, deploy.yml
```

#### Option 2: Manual

```bash
# Vào GitHub Actions → Deploy to Production → Run workflow
```

### Deploy to Staging

```bash
git push origin develop
# → Auto triggers deploy-staging.yml
```

### Run Database Migration

```bash
# Vào GitHub Actions → Database Migrations
# Select: environment (staging/production)
# Select: action (up/down/reset)
# → Run workflow
```

### Trigger Cleanup Job

```bash
# Manual:
# Vào GitHub Actions → Cleanup Expired Files → Run workflow

# Automatic:
# Chạy tự động mỗi ngày lúc 2 AM UTC
```

---

## 📊 Monitoring Workflow

### View Workflow Status

```
GitHub → Actions tab
```

### View Logs

```
GitHub → Actions → Select workflow → View logs
```

### Setup Slack Notifications

1. Tạo Slack Incoming Webhook:

   - Vào Slack → Apps → Incoming Webhooks
   - Add to channel
   - Copy webhook URL
2. Add vào GitHub Secrets:

   ```
   SLACK_WEBHOOK=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
   ```

---

## 🔍 Troubleshooting

### CI Tests Failed

```bash
# Xem logs trong GitHub Actions
# Fix code và push lại
git commit -am "fix: resolve test failures"
git push
```

### Build Failed

```bash
# Check Dockerfile
# Check dependencies trong go.mod
# Verify build locally:
docker build -t test .
```

### Deploy Failed

```bash
# Check SSH connection
ssh deploy@your-server.com

# Check Docker on server
docker ps
docker logs <container-id>

# Check GitHub Actions logs
```

### Migration Failed

```bash
# Rollback manually
ssh deploy@your-server.com
cd /opt/file-sharing-backend
psql -U postgres -d file_sharing_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Re-run migration workflow
```

---

## 📝 Best Practices

### Branch Strategy

```
main        → Production (protected)
  ↑
develop     → Staging (auto-deploy)
  ↑
feature/*   → Development (CI only)
```

### Release Process

```bash
# 1. Merge feature → develop
git checkout develop
git merge feature/new-feature
git push

# 2. Test in staging
# Wait for deploy-staging.yml to complete

# 3. Create PR: develop → main
# Get review & approval

# 4. Merge to main
# Creates deployment to production

# 5. Create release tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Rollback Strategy

```bash
# Option 1: Via GitHub (automatic on deploy failure)
# Deploy workflow has auto-rollback job

# Option 2: Manual rollback
ssh deploy@your-server.com
cd /opt/file-sharing-backend
docker-compose down
docker pull ghcr.io/yourusername/file-sharing-backend:v1.0.0-previous
docker-compose up -d
```

---

## 🔐 Security Checklist

- ✅ Never commit secrets to git
- ✅ Use GitHub Secrets for sensitive data
- ✅ Enable branch protection on main
- ✅ Require PR reviews before merge
- ✅ Use SSH key authentication (not password)
- ✅ Rotate secrets regularly
- ✅ Enable 2FA on GitHub
- ✅ Use environment-specific secrets
- ✅ Run security scans (gosec)
- ✅ Keep dependencies updated

---

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Nginx Documentation](https://nginx.org/en/docs/)

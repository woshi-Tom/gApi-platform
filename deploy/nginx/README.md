# Nginx Reverse Proxy Deployment Guide

## Overview

This guide covers deploying the gAPI Platform behind an Nginx reverse proxy with IPv4 + IPv6 dual-stack support.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (IPv4/IPv6)                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Nginx Reverse Proxy                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────┐ │
│  │  :80       │  │  :443      │  │  :8443      │  │ :8444  │ │
│  │  HTTP→HTTPS │  │  Consumer   │  │  Admin      │  │  API   │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
          ┌─────────────────────┼─────────────────────┐
          ▼                     ▼                     ▼
    ┌───────────┐        ┌───────────┐        ┌───────────┐
    │  :5173    │        │  :5174    │        │  :8080    │
    │  Consumer │        │   Admin   │        │  Backend  │
    │  (Vite)  │        │  (Vite)   │        │   (Go)    │
    └───────────┘        └───────────┘        └───────────┘
```

## Prerequisites

- Linux server (Ubuntu/Debian/CentOS/Alpine)
- Root or sudo access
- Frontend already built (`npm run build && npm run build:admin`)

## Quick Start

```bash
cd deploy/nginx
chmod +x deploy-nginx.sh
sudo ./deploy-nginx.sh --install
```

## Port Configuration

| Port | Protocol | Service | Path |
|------|----------|---------|------|
| 80 | HTTP | Redirect to HTTPS | / |
| 443 | HTTPS | Consumer App | / |
| 8443 | HTTPS | Admin Dashboard | /admin.html |
| 8444 | HTTPS | API Backend | /api/* |

## Configuration Files

### Main Configuration

```
/etc/nginx/conf.d/gapi-platform.conf
```

### SSL Certificates

```
/etc/nginx/ssl/gapi-platform.crt  (Certificate)
/etc/nginx/ssl/gapi-platform.key  (Private Key)
```

### Web Root

```
/var/www/gapi-platform/dist/        (Consumer App)
/var/www/gapi-platform/dist-admin/  (Admin Dashboard)
```

## Deployment Steps

### Step 1: Build Frontend

```bash
cd frontend
npm install
npm run build
npm run build:admin
```

### Step 2: Deploy with Script

```bash
cd deploy/nginx
chmod +x deploy-nginx.sh
sudo ./deploy-nginx.sh --install
```

The script will:
1. Install nginx (if not present)
2. Create necessary directories
3. Generate self-signed SSL certificates
4. Copy frontend build files
5. Install nginx configuration
6. Start nginx

### Step 3: Access the Application

Local access (self-signed cert):
- Consumer: https://localhost/ (accept certificate warning)
- Admin: https://localhost:8443/admin.html

Network access:
```bash
# Get IPv4 address
hostname -I | awk '{print $1}'

# Get IPv6 address
ip -6 addr show | grep global | awk '{print $2}' | cut -d'/' -f1
```

## Managing Nginx

### Reload Configuration

```bash
sudo ./deploy-nginx.sh --reload
```

### Test Configuration

```bash
sudo ./deploy-nginx.sh --test
```

### Check Status

```bash
sudo ./deploy-nginx.sh --status
```

### View Logs

```bash
# Access logs
sudo tail -f /var/log/nginx/access.log

# Error logs
sudo tail -f /var/log/nginx/error.log
```

### Uninstall

```bash
sudo ./deploy-nginx.sh --uninstall
```

## Production Deployment

### 1. Get Domain Name

Point your domain DNS to the server:
- A record: your-domain.com → server IPv4
- AAAA record: your-domain.com → server IPv6

### 2. Install SSL Certificate (Let's Encrypt)

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Obtain certificate (for consumer app)
sudo certbot --nginx -d your-domain.com

# For admin subdomain
sudo certbot --nginx -d admin.your-domain.com
```

### 3. Update Nginx Config

Edit `/etc/nginx/conf.d/gapi-platform.conf`:

```nginx
# Replace server_name _; with:
server_name your-domain.com;
```

### 4. Configure Firewall

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8443/tcp
sudo ufw allow 8444/tcp

# Firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8443/tcp
sudo firewall-cmd --permanent --add-port=8444/tcp
sudo firewall-cmd --reload
```

### 5. Set Up Auto-Renewal

```bash
# Test renewal
sudo certbot renew --dry-run

# Enable systemd timer (auto-renew)
sudo systemctl enable certbot-renew.timer
```

## Troubleshooting

### nginx Won't Start

```bash
# Check syntax
sudo nginx -t

# Check logs
sudo tail -50 /var/log/nginx/error.log
```

### Connection Refused

```bash
# Check if nginx is listening
sudo ss -tlnp | grep nginx

# Check firewall
sudo ufw status
```

### SSL Certificate Error

```bash
# Regenerate self-signed cert
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/nginx/ssl/gapi-platform.key \
    -out /etc/nginx/ssl/gapi-platform.crt \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=gAPI/OU=Dev/CN=localhost"
```

### IPv6 Not Working

```bash
# Check IPv6 support
ip -6 addr show

# Test IPv6 connectivity
curl -6 https://[your-ipv6]/api/v1/

# Check nginx listening on IPv6
sudo ss -tlnp | grep nginx
```

## Security Considerations

### Firewall

Ensure only necessary ports are open:
- 80 (HTTP redirect)
- 443 (HTTPS consumer)
- 8443 (HTTPS admin)
- 8444 (HTTPS API)

### SSL/TLS

- Use TLS 1.2+ only
- Disable weak ciphers
- Enable HSTS header
- Consider OCSP stapling

### Headers

The configuration includes security headers:
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Strict-Transport-Security

## File Structure

```
gapi-platform/
├── deploy/
│   └── nginx/
│       ├── gapi-platform.conf     # Nginx configuration
│       ├── deploy-nginx.sh        # Deployment script
│       └── README.md              # This file
└── frontend/
    ├── dist/                      # Built consumer app
    └── dist-admin/               # Built admin app
```

## Support

For issues:
1. Check nginx error logs
2. Verify all services are running
3. Test configuration with `nginx -t`
4. Ensure ports are not blocked by firewall

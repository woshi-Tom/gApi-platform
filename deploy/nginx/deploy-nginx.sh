#!/bin/bash
# =============================================================================
# gAPI Platform - Nginx Deployment Script
# =============================================================================
# Multi-distro compatible: Debian/Ubuntu, CentOS/RHEL, Alpine
# Supports IPv4 + IPv6 dual-stack
#
# Usage:
#   ./deploy-nginx.sh              # Interactive mode
#   ./deploy-nginx.sh --install   # Install and start
#   ./deploy-nginx.sh --uninstall # Remove nginx config
#   ./deploy-nginx.sh --reload    # Reload configuration
#   ./deploy-nginx.sh --test       # Test configuration
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory (resolve symlinks)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Configuration
NGINX_CONFIG_SOURCE="$SCRIPT_DIR/gapi-platform.conf"
NGINX_CONFIG_TARGET="/etc/nginx/conf.d/gapi-platform.conf"
SSL_CERT_DIR="/etc/nginx/ssl"
SSL_CERT="$SSL_CERT_DIR/gapi-platform.crt"
SSL_KEY="$SSL_CERT_DIR/gapi-platform.key"
WEB_ROOT="/var/www/gapi-platform"
FRONTEND_DIST="$PROJECT_ROOT/frontend/dist"
ADMIN_DIST="$PROJECT_ROOT/frontend/dist-admin"

# Commands
NGINX_CMD=""
NGINX_TEST_CMD=""
NGINX_RELOAD_CMD=""

# =============================================================================
# Utility Functions
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect Linux distribution
detect_distro() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        DISTRO="$ID"
        DISTRO_VERSION="$VERSION_ID"
    elif [[ -f /etc/centos-release ]]; then
        DISTRO="centos"
    elif [[ -f /etc/alpine-release ]]; then
        DISTRO="alpine"
    else
        DISTRO="unknown"
    fi
    
    case "$DISTRO" in
        ubuntu|debian)
            PKG_MANAGER="apt-get"
            PKG_INSTALL="apt-get install -y"
            ;;
        centos|rhel|rocky|almalinux)
            if command -v dnf &> /dev/null; then
                PKG_MANAGER="dnf"
                PKG_INSTALL="dnf install -y"
            else
                PKG_MANAGER="yum"
                PKG_INSTALL="yum install -y"
            fi
            ;;
        alpine)
            PKG_MANAGER="apk"
            PKG_INSTALL="apk add --no-cache"
            ;;
        *)
            log_warning "Unknown distribution, attempting apt-get..."
            PKG_MANAGER="apt-get"
            PKG_INSTALL="apt-get install -y"
            ;;
    esac
    
    log_info "Detected: $DISTRO ($PKG_MANAGER)"
}

# Check if command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# =============================================================================
# Installation Functions
# =============================================================================

install_nginx() {
    log_info "Installing nginx..."
    
    if command_exists nginx; then
        log_warning "nginx is already installed"
        return 0
    fi
    
    if command_exists $PKG_INSTALL; then
        $PKG_INSTALL nginx
    else
        log_error "Package manager not found. Please install nginx manually."
        exit 1
    fi
    
    if command_exists nginx; then
        log_success "nginx installed successfully"
    else
        log_error "Failed to install nginx"
        exit 1
    fi
}

create_directories() {
    log_info "Creating directories..."
    
    # Web root
    sudo mkdir -p "$WEB_ROOT"
    sudo mkdir -p "$WEB_ROOT/dist"
    sudo mkdir -p "$WEB_ROOT/dist-admin"
    
    # SSL directory
    sudo mkdir -p "$SSL_CERT_DIR"
    
    # Let's encrypt directory
    sudo mkdir -p /var/www/letsencrypt
    
    log_success "Directories created"
}

copy_ssl_certificates() {
    log_info "Setting up SSL certificates..."
    
    if [[ -f "$SSL_CERT" ]] && [[ -f "$SSL_KEY" ]]; then
        log_warning "SSL certificates already exist, skipping..."
        return 0
    fi
    
    # Generate self-signed certificate for local testing
    log_info "Generating self-signed SSL certificate for local testing..."
    
    sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$SSL_KEY" \
        -out "$SSL_CERT" \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=gAPI/OU=Dev/CN=localhost" \
        2>/dev/null
    
    sudo chmod 600 "$SSL_KEY"
    sudo chmod 644 "$SSL_CERT"
    
    log_success "Self-signed SSL certificate generated"
}

copy_frontend_files() {
    log_info "Copying frontend files..."
    
    # Copy consumer app
    if [[ -d "$FRONTEND_DIST" ]]; then
        sudo cp -r "$FRONTEND_DIST"/* "$WEB_ROOT/dist/"
        log_success "Consumer app copied"
    else
        log_warning "Frontend dist not found at $FRONTEND_DIST"
        log_warning "Please run 'npm run build' first"
    fi
    
    # Copy admin app
    if [[ -d "$ADMIN_DIST" ]]; then
        sudo cp -r "$ADMIN_DIST"/* "$WEB_ROOT/dist-admin/"
        log_success "Admin app copied"
    else
        log_warning "Admin dist not found at $ADMIN_DIST"
        log_warning "Please run 'npm run build:admin' first"
    fi
}

install_nginx_config() {
    log_info "Installing nginx configuration..."
    
    if [[ ! -f "$NGINX_CONFIG_SOURCE" ]]; then
        log_error "Nginx config not found at $NGINX_CONFIG_SOURCE"
        exit 1
    fi
    
    # Backup existing config if exists
    if [[ -f "$NGINX_CONFIG_TARGET" ]]; then
        sudo cp "$NGINX_CONFIG_TARGET" "${NGINX_CONFIG_TARGET}.backup.$(date +%Y%m%d%H%M%S)"
        log_warning "Backed up existing nginx config"
    fi
    
    # Copy new config
    sudo cp "$NGINX_CONFIG_SOURCE" "$NGINX_CONFIG_TARGET"
    
    log_success "Nginx configuration installed"
}

test_nginx_config() {
    log_info "Testing nginx configuration..."
    
    if ! $NGINX_TEST_CMD; then
        log_error "Nginx configuration test failed"
        $NGINX_TEST_CMD 2>&1 || true
        exit 1
    fi
    
    log_success "Configuration test passed"
}

start_nginx() {
    log_info "Starting nginx..."
    
    # Detect init system
    if command_exists systemctl; then
        sudo systemctl enable nginx
        sudo systemctl restart nginx
    elif command_exists service; then
        sudo service nginx restart
    elif [[ -f /etc/init.d/nginx ]]; then
        sudo /etc/init.d/nginx restart
    else
        log_warning "No init system detected, starting nginx manually..."
        sudo nginx
    fi
    
    # Verify nginx is running
    sleep 1
    if $NGINX_TEST_CMD 2>/dev/null; then
        log_success "nginx is running"
    else
        log_error "Failed to start nginx"
        exit 1
    fi
}

check_ports() {
    log_info "Checking listening ports..."
    
    echo ""
    echo "Ports that should be listening:"
    echo "  - Port 80  (HTTP redirect)"
    echo "  - Port 443 (HTTPS - Consumer App)"
    echo "  - Port 8443 (HTTPS - Admin App)"
    echo "  - Port 8444 (HTTPS - API Backend)"
    echo ""
    
    if command_exists ss; then
        echo "Current listening ports (with 'nginx' or '*:http'):"
        ss -tlnp | grep -E "nginx|:80|:443|:8443|:8444" || echo "  (no ports found yet)"
    elif command_exists netstat; then
        netstat -tlnp | grep -E "nginx|:80|:443|:8443|:8444" || echo "  (no ports found yet)"
    fi
}

show_access_info() {
    echo ""
    echo "============================================================================"
    log_success "gAPI Platform deployed successfully!"
    echo "============================================================================"
    echo ""
    
    # Get IPv6 addresses
    echo "Access URLs (Local Testing):"
    echo ""
    
    # IPv4
    IPV4=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "127.0.0.1")
    echo "  IPv4 (HTTP):"
    echo "    Consumer App:  http://$IPV4/"
    echo "    Admin App:    http://$IPV4:8443/admin.html"
    echo "    API Backend: http://$IPV4:8444/api/v1/"
    echo ""
    
    # IPv6
    IPV6=$(ip -6 addr show 2>/dev/null | grep "inet6" | grep "global" | awk '{print $2}' | cut -d'/' -f1 | head -1 || echo "")
    if [[ -n "$IPV6" ]]; then
        echo "  IPv6 (HTTPS):"
        echo "    Consumer App:  https://[$IPV6]/"
        echo "    Admin App:    https://[$IPV6]:8443/admin.html"
        echo "    API Backend: https://[$IPV6]:8444/api/v1/"
    else
        echo "  IPv6: No global IPv6 address detected"
    fi
    
    echo ""
    echo "Note: Self-signed SSL certificates require browser trust override."
    echo "      Visit each URL and accept the certificate warning."
    echo ""
    echo "============================================================================"
}

# =============================================================================
# Uninstallation
# =============================================================================

uninstall() {
    log_warning "Uninstalling nginx configuration..."
    
    if [[ -f "$NGINX_CONFIG_TARGET" ]]; then
        sudo rm "$NGINX_CONFIG_TARGET"
        log_success "Nginx configuration removed"
    fi
    
    if command_exists systemctl; then
        sudo systemctl reload nginx
    elif command_exists service; then
        sudo service nginx reload
    fi
    
    log_success "Uninstallation complete"
}

# =============================================================================
# Main
# =============================================================================

show_help() {
    cat << EOF
gAPI Platform - Nginx Deployment Script

Usage: $0 [COMMAND]

Commands:
    --install     Install and start nginx with gAPI configuration
    --uninstall  Remove nginx configuration
    --reload     Reload nginx configuration
    --test       Test nginx configuration
    --status     Show nginx status and ports
    --help       Show this help message

Examples:
    $0                    # Interactive mode
    $0 --install          # Install and start
    $0 --test             # Test config only
    $0 --reload           # Reload after config changes

EOF
}

main() {
    local ACTION="${1:-}"
    
    # Detect distribution
    detect_distro
    
    # Detect nginx commands
    if command_exists nginx; then
        NGINX_TEST_CMD="sudo nginx -t"
        NGINX_RELOAD_CMD="sudo nginx -s reload"
    fi
    
    case "$ACTION" in
        --install|-i)
            install_nginx
            create_directories
            create_ssl_certs
            copy_ssl_certificates
            copy_frontend_files
            install_nginx_config
            test_nginx_config
            start_nginx
            check_ports
            show_access_info
            ;;
        --uninstall|-u)
            uninstall
            ;;
        --reload|-r)
            if [[ -z "$NGINX_RELOAD_CMD" ]]; then
                log_error "nginx not found"
                exit 1
            fi
            test_nginx_config || exit 1
            $NGINX_RELOAD_CMD
            log_success "nginx configuration reloaded"
            ;;
        --test|-t)
            if [[ -z "$NGINX_TEST_CMD" ]]; then
                log_error "nginx not found"
                exit 1
            fi
            $NGINX_TEST_CMD
            ;;
        --status|-s)
            check_ports
            ;;
        --help|-h)
            show_help
            ;;
        "")
            # Interactive mode
            echo ""
            echo "gAPI Platform - Nginx Deployment"
            echo "================================"
            echo ""
            read -p "Do you want to install nginx and configure it for gAPI Platform? [Y/n]: " -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
                $0 --install
            else
                echo "Aborted."
                exit 0
            fi
            ;;
        *)
            log_error "Unknown option: $ACTION"
            show_help
            exit 1
            ;;
    esac
}

main "$@"

#!/bin/bash

# ==========================================================================
#  __  __       _ _        _         _ min
# |  \/  |     (_) |      / \   _ __| |__
# | |\/| | __ _| | |     / _ \ | '__| '_ \
# | |  | |/ _` | | |    / ___ \| |  | |_) |
# |_|  |_|\__,_|_|_|   /_/   \_\_|  |_.__/
#
# MailAdmin - Professional Native Installer & Updater
# ==========================================================================
# (c) 2026 Standalone Version. No Docker. No Cron.
# ==========================================================================

set -e

# --- Configuration ---
INSTALL_DIR="/opt/mailadmin"
GITHUB_REPO="https://github.com/kirpicheff/mailadmin.git"
SERVICE_NAME="mailadmin"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# --- UI Helpers ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC}  $1"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# --- Check Root ---
if [ "$EUID" -ne 0 ]; then
  error "This script must be run as root. Try: sudo $0"
fi

# --- Dependency Check ---
check_deps() {
    info "Checking system dependencies..."
    for cmd in git go npm node; do
        if ! command -v $cmd >/dev/null 2>&1; then
            error "$cmd is not installed. Please install it and try again."
        fi
    done
    ok "All dependencies found."
}

# --- Build & Restart Logic ---
build_and_restart() {
    info "Starting build process..."
    
    # Backend Build
    if [ -d "$INSTALL_DIR/backend" ]; then
        info "Building Go backend..."
        cd "$INSTALL_DIR/backend"
        go mod tidy
        go build -o "$INSTALL_DIR/mailadmin-bin" ./main.go
        ok "Backend binary created at $INSTALL_DIR/mailadmin-bin"
    else
        error "Backend directory not found in $INSTALL_DIR"
    fi

    # Frontend Build
    if [ -d "$INSTALL_DIR/frontend" ]; then
        info "Building Vue frontend (this may take a minute)..."
        cd "$INSTALL_DIR/frontend"
        npm install --no-audit --no-fund
        npm run build
        ok "Frontend built successfully in $INSTALL_DIR/frontend/dist"
    else
        error "Frontend directory not found in $INSTALL_DIR"
    fi

    # Service Management
    info "Configuring and restarting systemd service..."
    systemctl daemon-reload
    systemctl restart $SERVICE_NAME
    ok "Service ${SERVICE_NAME} is now running."
    
    status=$(systemctl is-active $SERVICE_NAME)
    if [ "$status" == "active" ]; then
        ok "MailAdmin is ACTIVE and running."
    else
        warn "Service started but status is: $status. Check logs: journalctl -u $SERVICE_NAME"
    fi
}

# --- Environment Configuration ---
setup_env() {
    ENV_FILE="$INSTALL_DIR/backend/.env"
    if [ ! -f "$ENV_FILE" ]; then
        info "Interactive configuration setup..."
        
        # DB_DSN
        DEFAULT_DSN="postfix:password@tcp(127.0.0.1:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
        printf "${YELLOW}Enter Database DSN${NC} [${DEFAULT_DSN}]: "
        read -r USER_DSN </dev/tty
        DB_DSN=${USER_DSN:-$DEFAULT_DSN}

        # JWT_SECRET
        DEFAULT_JWT=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 32)
        printf "${YELLOW}Enter JWT Secret${NC} (blank for random): "
        read -r USER_JWT </dev/tty
        JWT_SECRET=${USER_JWT:-$DEFAULT_JWT}

        # LISTEN_ADDR
        DEFAULT_LISTEN=":8080"
        printf "${YELLOW}Enter API Port${NC} [${DEFAULT_LISTEN}]: "
        read -r USER_LISTEN </dev/tty
        LISTEN_ADDR=${USER_LISTEN:-$DEFAULT_LISTEN}

        # Save to file
        cat <<EOF > "$ENV_FILE"
DB_DSN=$DB_DSN
JWT_SECRET=$JWT_SECRET
LISTEN_ADDR=$LISTEN_ADDR
EOF
        chmod 600 "$ENV_FILE"
        ok "Environment file created and secured at $ENV_FILE"
    else
        info "Using existing configuration at $ENV_FILE"
    fi
}

# --- Systemd Setup ---
setup_service() {
    if [ ! -f "$SERVICE_FILE" ]; then
        info "Installing systemd unit file..."
        cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=MailAdmin Native Backend
After=network.target mariadb.service mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR/backend
ExecStart=$INSTALL_DIR/mailadmin-bin
Restart=always
EnvironmentFile=$INSTALL_DIR/backend/.env

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable $SERVICE_NAME
        ok "Systemd service registered."
    fi
}

# --- Main Entry Point ---
echo -e "${BLUE}====================================================${NC}"
echo -e "${BLUE}   MailAdmin Installation & Update Manager${NC}"
echo -e "${BLUE}====================================================${NC}"

if [ -d "$INSTALL_DIR/.git" ]; then
    # --- UPDATE FLOW ---
    info "Detected existing installation in $INSTALL_DIR"
    cd "$INSTALL_DIR"
    
    info "Checking for updates on GitHub..."
    git fetch origin main
    
    LOCAL_HASH=$(git rev-parse HEAD)
    REMOTE_HASH=$(git rev-parse origin/main)
    
    if [ "$LOCAL_HASH" != "$REMOTE_HASH" ]; then
        warn "Update available! Current: ${LOCAL_HASH:0:7}, Remote: ${REMOTE_HASH:0:7}"
        read -p "Do you want to update now? [Y/n]: " DO_UPDATE </dev/tty
        if [[ ! "$DO_UPDATE" =~ ^[Nn]$ ]]; then
            info "Updating source code..."
            git pull origin main
            setup_env
            build_and_restart
            ok "MailAdmin updated to the latest version."
        else
            info "Update cancelled."
        fi
    else
        ok "You are already on the latest version."
        read -p "Would you like to force a rebuild? [y/N]: " FORCE </dev/tty
        if [[ "$FORCE" =~ ^[Yy]$ ]]; then
            build_and_restart
        fi
    fi
else
    # --- INSTALL FLOW ---
    info "Installing MailAdmin to $INSTALL_DIR..."
    check_deps
    
    if [ ! -d "$INSTALL_DIR" ]; then
        git clone "$GITHUB_REPO" "$INSTALL_DIR"
        ok "Repository cloned."
    fi
    
    cd "$INSTALL_DIR"
    setup_env
    setup_service
    build_and_restart
    
    echo -e "${GREEN}====================================================${NC}"
    ok "INSTALLATION FINISHED!"
    info "Web Port: 8081 (Check docker-compose.yml or your Nginx config)"
    info "API Port: $LISTEN_ADDR"
    warn "Make sure to point your Nginx to $INSTALL_DIR/frontend/dist"
    echo -e "${GREEN}====================================================${NC}"
fi

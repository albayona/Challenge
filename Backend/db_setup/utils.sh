#!/bin/bash

# Database Setup Utilities
# Common functions for all database setup scripts
# Follows .cursorrules standards: clean, focused, error handling

# Colors for output - following descriptive naming
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging functions - small, focused functions
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

# Load environment variables from .env file safely
load_env_file() {
    local env_file="${1:-../.env}"
    
    if [ -f "$env_file" ]; then
        log_info "Loading environment variables from $env_file"
        # Use a safer method to load .env file
        while IFS= read -r line; do
            # Skip comments and empty lines
            if [[ $line =~ ^[[:space:]]*# ]] || [[ -z "${line// }" ]]; then
                continue
            fi
            # Export the variable safely
            export "$line" 2>/dev/null || true
        done < "$env_file"
    else
        log_warning "Environment file $env_file not found"
    fi
}

# Check if a port is in use
check_port() {
    local port="$1"
    if lsof -i ":$port" >/dev/null 2>&1; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Kill processes on a specific port
kill_port() {
    local port="$1"
    local pids=$(lsof -ti ":$port" 2>/dev/null || true)
    
    if [ -n "$pids" ]; then
        log_info "Killing processes on port $port: $pids"
        echo "$pids" | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
}

# Check if CockroachDB is running
check_cockroach_running() {
    if pgrep -f cockroach >/dev/null 2>&1; then
        return 0  # CockroachDB is running
    else
        return 1  # CockroachDB is not running
    fi
}

# Wait for CockroachDB to be ready
wait_for_cockroach() {
    local host="${1:-localhost}"
    local port="${2:-26257}"
    local max_attempts="${3:-30}"
    local attempt=0
    
    log_info "Waiting for CockroachDB to be ready on $host:$port"
    
    while [ $attempt -lt $max_attempts ]; do
        if cockroach sql --host="$host:$port" --certs-dir="../db_setup_files/certs" --execute="SELECT 1;" >/dev/null 2>&1; then
            log_success "CockroachDB is ready"
            return 0
        fi
        
        attempt=$((attempt + 1))
        log_info "Attempt $attempt/$max_attempts - waiting for CockroachDB..."
        sleep 2
    done
    
    log_error "CockroachDB failed to become ready after $max_attempts attempts"
    return 1
}

# Create database if it doesn't exist
create_database_if_not_exists() {
    local db_name="${1:-stock_data}"
    local host="${2:-localhost}"
    local port="${3:-26257}"
    local certs_dir="${4:-../db_setup_files/certs}"
    
    log_info "Creating database '$db_name' if it doesn't exist"
    
    if cockroach sql --host="$host:$port" --certs-dir="$certs_dir" --execute="CREATE DATABASE IF NOT EXISTS $db_name;" >/dev/null 2>&1; then
        log_success "Database '$db_name' is ready"
        return 0
    else
        log_error "Failed to create database '$db_name'"
        return 1
    fi
}

# Check if database exists
database_exists() {
    local db_name="${1:-stock_data}"
    local host="${2:-localhost}"
    local port="${3:-26257}"
    local certs_dir="${4:-../db_setup_files/certs}"
    
    if cockroach sql --host="$host:$port" --certs-dir="$certs_dir" --execute="SELECT 1 FROM information_schema.schemata WHERE schema_name = '$db_name';" 2>/dev/null | grep -q "1 row"; then
        return 0  # Database exists
    else
        return 1  # Database doesn't exist
    fi
}

# Get database configuration
get_db_config() {
    # Default values
    local host="${COCKROACH_HOST:-localhost}"
    local port="${COCKROACH_PORT:-26257}"
    local user="${COCKROACH_USER:-root}"
    local db_name="${COCKROACH_DB_NAME:-stock_data}"
    local ssl_mode="${COCKROACH_SSL_MODE:-require}"
    local certs_dir="../db_setup_files/certs"
    
    echo "host=$host"
    echo "port=$port"
    echo "user=$user"
    echo "db_name=$db_name"
    echo "ssl_mode=$ssl_mode"
    echo "certs_dir=$certs_dir"
}

# Clean up temporary files
cleanup_temp_files() {
    log_info "Cleaning up temporary files"
    rm -f node1.pid node2.pid node3.pid
    rm -f cockroach-temp*
    rm -f *.log
}

# Validate required environment variables
validate_env_vars() {
    local required_vars=("COCKROACH_HOST" "COCKROACH_PORT" "COCKROACH_USER" "COCKROACH_DB_NAME")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var:-}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        log_error "Missing required environment variables: ${missing_vars[*]}"
        return 1
    fi
    
    return 0
}

# Print script header
print_script_header() {
    local script_name="$1"
    local description="$2"
    
    echo "=========================================="
    echo "  $script_name"
    echo "  $description"
    echo "=========================================="
    echo
}

# Print script footer
print_script_footer() {
    local message="${1:-Script completed successfully}"
    echo
    echo "=========================================="
    log_success "$message"
    echo "=========================================="
}

# Check if CockroachDB is installed
check_cockroachdb_installed() {
    if ! command -v cockroach &> /dev/null; then
        log_error "CockroachDB is not installed!"
        echo "Please install CockroachDB first:"
        echo "https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html"
        exit 1
    fi
    log_success "CockroachDB is installed"
}

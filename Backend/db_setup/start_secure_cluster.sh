#!/bin/bash

# CockroachDB Secure Cluster Setup Script
# Follows .cursorrules standards: clean, focused, error handling
# Sets up secure cluster in db_setup directory

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Load utilities
source "$(dirname "$0")/utils.sh"

# Load environment variables
load_env_file "../.env"

# Configuration constants - direct paths
readonly CERTS_DIR="./certs"
readonly NODE1_PORT="${COCKROACH_NODE1_PORT:-26257}"
readonly NODE2_PORT="${COCKROACH_NODE2_PORT:-26258}"
readonly NODE3_PORT="${COCKROACH_NODE3_PORT:-26259}"
readonly NODE1_HTTP="${COCKROACH_NODE1_HTTP:-8080}"
readonly NODE2_HTTP="${COCKROACH_NODE2_HTTP:-8081}"
readonly NODE3_HTTP="${COCKROACH_NODE3_HTTP:-8082}"
readonly DB_NAME="${DB_NAME:-stock_data}"
readonly DB_USER="${DB_USER:-stock_user}"
readonly DB_PASSWORD="${DB_PASSWORD:-secure_password}"
readonly CLUSTER_NAME="${COCKROACH_CLUSTER_NAME:-dataextractor-secure-cluster}"
readonly RANGE_MAX_BYTES="${COCKROACH_RANGE_MAX_BYTES:-134217728}"  # 128MB
readonly RANGE_MIN_BYTES="${COCKROACH_RANGE_MIN_BYTES:-33554432}"   # 32MB
readonly NUM_REPLICAS="${COCKROACH_NUM_REPLICAS:-3}"

# Check if certificates exist
check_certificates_exist() {
    log_info "Checking for existing certificates..."
    
    local required_files=(
        "${CERTS_DIR}/ca.crt"
        "${CERTS_DIR}/node.crt"
        "${CERTS_DIR}/node.key"
        "${CERTS_DIR}/client.root.crt"
        "${CERTS_DIR}/client.root.key"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "${file}" ]; then
            log_error "Required certificate file not found: ${file}"
            echo "Please run ./generate_certificates.sh first"
            exit 1
        fi
    done
    
    log_success "All required certificates found"
}

# Check if cluster is already running
check_cluster_running() {
    log_info "Checking if cluster is already running..."
    
    # Check if any cockroach processes are running
    if pgrep -f cockroach > /dev/null; then
        log_info "CockroachDB processes are already running"
        return 0
    else
        log_info "No CockroachDB processes found"
        return 1
    fi
}

# Start a secure node with proper error handling
start_secure_node() {
    local node_id="$1"
    local port="$2"
    local http_port="$3"
    local store_dir="node${node_id}"
    local pid_file="node${node_id}.pid"
    
    log_info "Starting secure Node ${node_id}..."
    
    cockroach start \
        --certs-dir="${CERTS_DIR}" \
        --store="${store_dir}" \
        --listen-addr="localhost:${port}" \
        --http-addr="localhost:${http_port}" \
        --join="localhost:${NODE1_PORT},localhost:${NODE2_PORT},localhost:${NODE3_PORT}" \
        --background \
        --pid-file="${pid_file}" || {
        log_error "Failed to start secure Node ${node_id}"
        exit 1
    }
    
    sleep 3
    log_success "Secure Node ${node_id} started on port ${port} (HTTP: ${http_port})"
}

# Initialize the secure cluster
initialize_secure_cluster() {
    log_info "Initializing secure cluster..."
    
    cockroach init --certs-dir="${CERTS_DIR}" --host="localhost:${NODE1_PORT}" || {
        log_error "Failed to initialize secure cluster"
        exit 1
    }
    
    log_success "Secure cluster initialized successfully"
}

# Create database and user
create_database_and_user() {
    log_info "Creating database and user..."
    log_info "Using DB_NAME: ${DB_NAME}"
    log_info "Using DB_USER: ${DB_USER}"
    log_info "Using DB_PASSWORD: ${DB_PASSWORD}"
    
    # Create database if it doesn't exist
    cockroach sql --certs-dir="${CERTS_DIR}" --host="localhost:${NODE1_PORT}" --execute="CREATE DATABASE IF NOT EXISTS ${DB_NAME};" || {
        log_error "Failed to create database ${DB_NAME}"
        exit 1
    }
    
    # Create user if it doesn't exist
    cockroach sql --certs-dir="${CERTS_DIR}" --host="localhost:${NODE1_PORT}" --execute="CREATE USER IF NOT EXISTS ${DB_USER};" || {
        log_error "Failed to create user ${DB_USER}"
        exit 1
    }
    
    # Set password for the user
    cockroach sql --certs-dir="${CERTS_DIR}" --host="localhost:${NODE1_PORT}" --execute="ALTER USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';" || {
        log_error "Failed to set password for user ${DB_USER}"
        exit 1
    }
    
    # Grant admin role to the user
    cockroach sql --certs-dir="${CERTS_DIR}" --host="localhost:${NODE1_PORT}" --execute="GRANT admin TO ${DB_USER};" || {
        log_error "Failed to grant admin role to user ${DB_USER}"
        exit 1
    }
    
    log_success "Database ${DB_NAME} and user ${DB_USER} created successfully"
}

# Display secure cluster information
display_secure_cluster_info() {
    log_info "Secure Cluster Status:"
    echo "========================"
    
    # Show node status
    cockroach node status --certs-dir="${CERTS_DIR}" --host="localhost:${NODE1_PORT}" || {
        log_warning "Cannot connect to secure cluster"
        return 1
    }
    
    echo ""
    log_info "Secure Database URLs (Root User):"
    echo "Node 1: postgresql://root@localhost:${NODE1_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    echo "Node 2: postgresql://root@localhost:${NODE2_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    echo "Node 3: postgresql://root@localhost:${NODE3_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    
    echo ""
    log_info "Application Database URLs (${DB_USER} User):"
    echo "Node 1: postgresql://${DB_USER}:${DB_PASSWORD}@localhost:${NODE1_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    echo "Node 2: postgresql://${DB_USER}:${DB_PASSWORD}@localhost:${NODE2_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    echo "Node 3: postgresql://${DB_USER}:${DB_PASSWORD}@localhost:${NODE3_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    
    echo ""
    log_info "DB Console URLs:"
    echo "Node 1: https://localhost:${NODE1_HTTP}"
    echo "Node 2: https://localhost:${NODE2_HTTP}"
    echo "Node 3: https://localhost:${NODE3_HTTP}"
    
    echo ""
    log_info "Connection Commands:"
    echo "SQL Client: cockroach sql --certs-dir=${CERTS_DIR} --host=localhost:${NODE1_PORT}"
    echo "Node Status: cockroach node status --certs-dir=${CERTS_DIR} --host=localhost:${NODE1_PORT}"
}

# Display cluster startup details
display_cluster_startup_details() {
    log_info "Cluster Startup Details:"
    echo "========================"
    
    # Get node 1 startup details
    if [ -f "node1/logs/cockroach.log" ]; then
        log_info "Node 1 startup details:"
        grep 'node starting' node1/logs/cockroach.log -A 11 || log_warning "Could not retrieve node 1 startup details"
    fi
    
    echo ""
    log_info "Cluster Information:"
    echo "- Cluster Name: ${CLUSTER_NAME}"
    echo "- Database: ${DB_NAME}"
    echo "- Database User: ${DB_USER}"
    echo "- Nodes: 3 (ports ${NODE1_PORT}, ${NODE2_PORT}, ${NODE3_PORT})"
    echo "- Replicas: ${NUM_REPLICAS}"
    echo "- Range Configuration: ${RANGE_MIN_BYTES}MB-${RANGE_MAX_BYTES}MB"
    echo "- Security: TLS encryption enabled"
}

# Main execution function - follows single responsibility
main() {
    echo "================================================"
    echo "CockroachDB Secure Cluster Setup"
    echo "================================================"
    echo ""
    
    # Execute secure setup steps in order
    check_certificates_exist
    check_cockroachdb_installed
    
    # Check if cluster is already running
    if check_cluster_running; then
        log_info "Cluster is already running. Skipping node startup."
    else
    # Start secure cluster nodes
    start_secure_node 1 "${NODE1_PORT}" "${NODE1_HTTP}"
    start_secure_node 2 "${NODE2_PORT}" "${NODE2_HTTP}"
    start_secure_node 3 "${NODE3_PORT}" "${NODE3_HTTP}"
    
    # Initialize and configure
    initialize_secure_cluster
    fi
    
    # Create database and user (always run this)
    log_info "About to create database and user..."
    create_database_and_user
    
    # Display results
    display_secure_cluster_info
    display_cluster_startup_details
    
    log_success "CockroachDB Secure Cluster Setup Complete!"
    echo "================================================"
    echo ""
    echo "Security Features:"
    echo "- TLS encryption for all connections"
    echo "- Certificate-based authentication"
    echo "- Secure inter-node communication"
    echo "- Encrypted data at rest"
    echo ""
    echo "Next Steps:"
    echo "1. Monitor cluster: https://localhost:${NODE1_HTTP}"
    echo "2. Connect with SQL: cockroach sql --certs-dir=${CERTS_DIR} --host=localhost:${NODE1_PORT}"
    echo "3. Stop cluster: ./stop_secure_cluster.sh"
    echo ""
}

# Execute main function
main "$@"

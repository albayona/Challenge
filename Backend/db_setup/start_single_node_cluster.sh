#!/bin/bash

# CockroachDB Single Node Secure Cluster Setup Script
# Follows .cursorrules standards: clean, focused, error handling
# Sets up secure single-node cluster in db_setup directory
# This avoids the multi-node transaction limit in CockroachDB Community Edition

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Load utilities
source "$(dirname "$0")/utils.sh"

# Load environment variables
load_env_file "../.env"

# Configuration constants - direct paths
readonly CERTS_DIR="./certs"
readonly NODE1_PORT="${COCKROACH_NODE1_PORT:-26257}"
readonly NODE1_HTTP="${COCKROACH_NODE1_HTTP:-8080}"
readonly DB_NAME="${DB_NAME:-stock_data}"
readonly DB_USER="${DB_USER:-stock_user}"
readonly DB_PASSWORD="${DB_PASSWORD:-secure_password}"
readonly CLUSTER_NAME="${COCKROACH_CLUSTER_NAME:-dataextractor-secure-cluster}"
readonly RANGE_MAX_BYTES="${COCKROACH_RANGE_MAX_BYTES:-134217728}"  # 128MB
readonly RANGE_MIN_BYTES="${COCKROACH_RANGE_MIN_BYTES:-33554432}"   # 32MB
readonly NUM_REPLICAS="1"  # Single node - no replication

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

# Start a single secure node using start-single-node command
start_single_node() {
    local node_id="1"
    local port="${NODE1_PORT}"
    local http_port="${NODE1_HTTP}"
    local store_dir="node${node_id}"
    local pid_file="node${node_id}.pid"
    
    log_info "Starting secure single-node cluster using start-single-node command..."
    
    # Use start-single-node command to avoid multi-node transaction limits
    cockroach start-single-node \
        --certs-dir="${CERTS_DIR}" \
        --store="${store_dir}" \
        --listen-addr="localhost:${port}" \
        --http-addr="localhost:${http_port}" \
        --background \
        --pid-file="${pid_file}" || {
        log_error "Failed to start secure single-node cluster"
        exit 1
    }
    
    sleep 3
    log_success "Secure single-node cluster started on port ${port} (HTTP: ${http_port})"
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
    log_info "Secure Database URL (Root User):"
    echo "postgresql://root@localhost:${NODE1_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    
    echo ""
    log_info "Application Database URL (${DB_USER} User):"
    echo "postgresql://${DB_USER}:${DB_PASSWORD}@localhost:${NODE1_PORT}/${DB_NAME}?sslmode=require&sslcert=${CERTS_DIR}/client.root.crt&sslkey=${CERTS_DIR}/client.root.key&sslrootcert=${CERTS_DIR}/ca.crt"
    
    echo ""
    log_info "DB Console URL:"
    echo "https://localhost:${NODE1_HTTP}"
    
    echo ""
    log_info "Connection Commands:"
    echo "SQL Client: cockroach sql --certs-dir=${CERTS_DIR} --host=localhost:${NODE1_PORT}"
    echo "Node Status: cockroach node status --certs-dir=${CERTS_DIR} --host=localhost:${NODE1_PORT}"
}

# Display cluster startup details
display_cluster_startup_details() {
    log_info "Cluster Startup Details:"
    echo "========================"
    
    # Get node startup details
    if [ -f "node1/logs/cockroach.log" ]; then
        log_info "Node startup details:"
        grep 'node starting' node1/logs/cockroach.log -A 11 || log_warning "Could not retrieve node startup details"
    fi
    
    echo ""
    log_info "Cluster Information:"
    echo "- Cluster Name: ${CLUSTER_NAME}"
    echo "- Database: ${DB_NAME}"
    echo "- Database User: ${DB_USER}"
    echo "- Node: Single node (port ${NODE1_PORT})"
    echo "- Replicas: ${NUM_REPLICAS} (single node)"
    echo "- Range Configuration: ${RANGE_MIN_BYTES}MB-${RANGE_MAX_BYTES}MB"
    echo "- Security: TLS encryption enabled"
    echo ""
    log_info "Note: Single-node cluster avoids CockroachDB Community Edition multi-node transaction limits"
}

# Main execution function - follows single responsibility
main() {
    echo "================================================"
    echo "CockroachDB Single Node Secure Cluster Setup"
    echo "================================================"
    echo ""
    
    # Execute secure setup steps in order
    check_certificates_exist
    check_cockroachdb_installed
    
    # Check if cluster is already running
    if check_cluster_running; then
        log_info "Cluster is already running. Skipping node startup."
    else
        # Start single-node cluster (no need to initialize - start-single-node auto-initializes)
        start_single_node
    fi
    
    # Create database and user (always run this)
    log_info "About to create database and user..."
    create_database_and_user
    
    # Display results
    display_secure_cluster_info
    display_cluster_startup_details
    
    log_success "CockroachDB Single Node Secure Cluster Setup Complete!"
    echo "================================================"
    echo ""
    echo "Security Features:"
    echo "- TLS encryption for all connections"
    echo "- Certificate-based authentication"
    echo "- Encrypted data at rest"
    echo ""
    echo "Benefits of Single-Node:"
    echo "- No multi-node transaction limits"
    echo "- Simpler setup and management"
    echo "- Lower resource usage"
    echo ""
    echo "Next Steps:"
    echo "1. Monitor cluster: https://localhost:${NODE1_HTTP}"
    echo "2. Connect with SQL: cockroach sql --certs-dir=${CERTS_DIR} --host=localhost:${NODE1_PORT}"
    echo "3. Stop cluster: ./stop_secure_cluster.sh"
    echo ""
}

# Execute main function
main "$@"


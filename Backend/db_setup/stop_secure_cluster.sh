#!/bin/bash

# CockroachDB Secure Cluster Stop Script
# Gracefully stops all running CockroachDB nodes
# Follows .cursorrules standards: clean, focused, error handling

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Load utilities
source "$(dirname "$0")/utils.sh"

# Load environment variables
load_env_file "../.env"

# Configuration constants
readonly CERTS_DIR="./certs"
readonly NODE1_PORT="${COCKROACH_NODE1_PORT:-26257}"
readonly NODE2_PORT="${COCKROACH_NODE2_PORT:-26258}"
readonly NODE3_PORT="${COCKROACH_NODE3_PORT:-26259}"

# Get all running CockroachDB processes
get_cockroach_processes() {
    log_info "Finding running CockroachDB processes..."
    
    # Get process IDs of running cockroach processes
    local pids=$(ps -ef | grep cockroach | grep -v grep | awk '{print $2}' || true)
    
    if [ -z "$pids" ]; then
        log_info "No CockroachDB processes found"
        return 1
    fi
    
    echo "$pids"
    return 0
}

# Gracefully stop a single node
stop_node() {
    local pid="$1"
    local node_info="$2"
    
    log_info "Stopping node (PID: ${pid}) - ${node_info}"
    
    # Send TERM signal for graceful shutdown
    if kill -TERM "$pid" 2>/dev/null; then
        log_info "TERM signal sent to PID ${pid}"
        
        # Wait for graceful shutdown (max 30 seconds)
        local count=0
        while kill -0 "$pid" 2>/dev/null && [ $count -lt 30 ]; do
            sleep 1
            count=$((count + 1))
        done
        
        # Check if process is still running
        if kill -0 "$pid" 2>/dev/null; then
            log_warning "Process ${pid} did not stop gracefully, forcing shutdown..."
            kill -KILL "$pid" 2>/dev/null || true
        else
            log_success "Node stopped gracefully (PID: ${pid})"
        fi
    else
        log_warning "Could not send TERM signal to PID ${pid} (process may have already stopped)"
    fi
}

# Stop all CockroachDB nodes
stop_all_nodes() {
    log_info "Stopping all CockroachDB nodes..."
    
    # Get all running processes
    local pids
    if ! pids=$(get_cockroach_processes); then
        log_info "No CockroachDB processes to stop"
        return 0
    fi
    
    # Display running processes
    log_info "Found running CockroachDB processes:"
    ps -ef | grep cockroach | grep -v grep || true
    
    # Stop each process
    for pid in $pids; do
        # Get process details for logging
        local process_info=$(ps -p "$pid" -o args= 2>/dev/null || echo "Unknown process")
        stop_node "$pid" "$process_info"
    done
    
    log_success "All CockroachDB nodes stopped"
}

# Display cluster status after shutdown
display_shutdown_status() {
    log_info "Checking cluster status after shutdown..."
    
    # Check if any processes are still running
    if ps -ef | grep cockroach | grep -v grep > /dev/null; then
        log_warning "Some CockroachDB processes may still be running:"
        ps -ef | grep cockroach | grep -v grep || true
    else
        log_success "All CockroachDB processes have been stopped"
    fi
    
    echo ""
    log_info "Cluster shutdown complete!"
    echo "================================================"
    echo ""
    echo "To restart the cluster:"
    echo "  ./start_secure_cluster.sh"
    echo ""
    echo "Note: Data directories are preserved and can be restarted"
}

# Main execution function
main() {
    echo "================================================"
    echo "CockroachDB Secure Cluster Shutdown"
    echo "================================================"
    echo ""
    
    # Stop all nodes gracefully
    stop_all_nodes
    
    # Display final status
    display_shutdown_status
}

# Execute main function
main "$@"

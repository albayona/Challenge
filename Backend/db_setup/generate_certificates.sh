#!/bin/bash

# CockroachDB Certificate Generation Script
# Follows .cursorrules standards: clean, focused, error handling
# Generates certificates in db_setup/certs directory

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Load utilities
source "$(dirname "$0")/utils.sh"

# Load environment variables
load_env_file "../.env"

# Configuration constants - direct paths
readonly CERTS_DIR="./certs"
readonly SAFE_DIR="./my-safe-directory"
readonly CA_KEY_FILE="${SAFE_DIR}/ca.key"
readonly ROOT_USER="${COCKROACH_USER:-root}"
readonly NODE_HOSTNAME="${COCKROACH_HOST:-localhost}"
readonly HOSTNAME=$(hostname)

# Clean up existing certificates
cleanup_existing_certificates() {
    log_info "Cleaning up existing certificates..."
    
    # Remove certificate directories if they exist
    if [ -d "${CERTS_DIR}" ]; then
        log_info "Removing existing certs directory..."
        rm -rf "${CERTS_DIR}"
    fi
    
    if [ -d "${SAFE_DIR}" ]; then
        log_info "Removing existing safe directory..."
        rm -rf "${SAFE_DIR}"
    fi
    
    log_success "Cleanup completed"
}

# Create certificate directories
create_certificate_directories() {
    log_info "Creating certificate directories..."
    
    # Create certs directory
    mkdir -p "${CERTS_DIR}" || {
        log_error "Failed to create ${CERTS_DIR} directory"
        exit 1
    }
    
    # Create safe directory
    mkdir -p "${SAFE_DIR}" || {
        log_error "Failed to create ${SAFE_DIR} directory"
        exit 1
    }
    
    log_success "Certificate directories created"
}

# Generate CA certificate and key
generate_ca_certificate() {
    log_info "Generating CA certificate and key..."
    
    cockroach cert create-ca \
        --certs-dir="${CERTS_DIR}" \
        --ca-key="${CA_KEY_FILE}" || {
        log_error "Failed to create CA certificate"
        exit 1
    }
    
    log_success "CA certificate and key generated"
}

# Generate node certificate
generate_node_certificate() {
    log_info "Generating node certificate for ${NODE_HOSTNAME} and ${HOSTNAME}..."
    
    cockroach cert create-node \
        "${NODE_HOSTNAME}" \
        "${HOSTNAME}" \
        --certs-dir="${CERTS_DIR}" \
        --ca-key="${CA_KEY_FILE}" || {
        log_error "Failed to create node certificate"
        exit 1
    }
    
    log_success "Node certificate generated"
}

# Generate client certificate for root user
generate_client_certificate() {
    log_info "Generating client certificate for ${ROOT_USER}..."
    
    cockroach cert create-client \
        "${ROOT_USER}" \
        --certs-dir="${CERTS_DIR}" \
        --ca-key="${CA_KEY_FILE}" || {
        log_error "Failed to create client certificate"
        exit 1
    }
    
    log_success "Client certificate for ${ROOT_USER} generated"
}

# Verify certificate generation
verify_certificates() {
    log_info "Verifying generated certificates..."
    
    # Check if required files exist
    local required_files=(
        "${CERTS_DIR}/ca.crt"
        "${CERTS_DIR}/node.crt"
        "${CERTS_DIR}/node.key"
        "${CERTS_DIR}/client.root.crt"
        "${CERTS_DIR}/client.root.key"
        "${CA_KEY_FILE}"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "${file}" ]; then
            log_error "Required certificate file not found: ${file}"
            exit 1
        fi
    done
    
    log_success "All required certificates verified"
}

# Display certificate information
display_certificate_info() {
    log_info "Certificate Information:"
    echo "========================"
    
    echo ""
    log_info "Generated Files:"
    echo "CA Certificate: ${CERTS_DIR}/ca.crt"
    echo "CA Key: ${CERTS_DIR}/ca.key"
    echo "Node Certificate: ${CERTS_DIR}/node.crt"
    echo "Node Key: ${CERTS_DIR}/node.key"
    echo "Client Certificate: ${CERTS_DIR}/client.root.crt"
    echo "Client Key: ${CERTS_DIR}/client.root.key"
    echo "CA Key (Safe): ${CA_KEY_FILE}"
    
    echo ""
    log_info "Certificate Details:"
    
    # Display CA certificate info
    echo "CA Certificate:"
    openssl x509 -in "${CERTS_DIR}/ca.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After)" || true
    
    echo ""
    echo "Node Certificate:"
    openssl x509 -in "${CERTS_DIR}/node.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After)" || true
    
    echo ""
    echo "Client Certificate:"
    openssl x509 -in "${CERTS_DIR}/client.root.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After)" || true
}

# Set proper permissions on certificate files
set_certificate_permissions() {
    log_info "Setting proper permissions on certificate files..."
    
    # Set restrictive permissions on private keys
    chmod 600 "${CERTS_DIR}/ca.key" "${CERTS_DIR}/node.key" "${CERTS_DIR}/client.root.key" "${CA_KEY_FILE}" || {
        log_warning "Failed to set permissions on private keys"
    }
    
    # Set readable permissions on certificates
    chmod 644 "${CERTS_DIR}/ca.crt" "${CERTS_DIR}/node.crt" "${CERTS_DIR}/client.root.crt" || {
        log_warning "Failed to set permissions on certificates"
    }
    
    log_success "Certificate permissions set"
}

# Create environment file for secure connections
create_secure_environment() {
    log_info "Creating secure environment configuration..."
    
    cat > ../.env.secure << EOF
# Secure CockroachDB Configuration
COCKROACH_HOST=localhost
COCKROACH_PORT=26257
COCKROACH_USER=root
COCKROACH_PASSWORD=
COCKROACH_DB_NAME=dataextractor
COCKROACH_SSL_MODE=require
COCKROACH_CERTS_DIR=${CERTS_DIR}

# Certificate paths
COCKROACH_CA_CERT=${CERTS_DIR}/ca.crt
COCKROACH_CLIENT_CERT=${CERTS_DIR}/client.root.crt
COCKROACH_CLIENT_KEY=${CERTS_DIR}/client.root.key

# API Configuration
API_BASE_URL=https://your-api-url.com
API_KEY=your-api-key-here
API_ENDPOINT=/data
OUTPUT_FILE=extracted_data.json
EOF
    
    log_success "Secure environment file created at ../.env.secure"
}

# Display usage instructions
display_usage_instructions() {
    log_success "Certificate Generation Complete!"
    echo "====================================="
    echo ""
    echo "Generated Certificates:"
    echo "- CA Certificate: ${CERTS_DIR}/ca.crt"
    echo "- Node Certificate: ${CERTS_DIR}/node.crt"
    echo "- Client Certificate: ${CERTS_DIR}/client.root.crt"
    echo "- CA Key (Safe): ${CA_KEY_FILE}"
    echo ""
    echo "Next Steps:"
    echo "1. Start secure cluster: cockroach start --certs-dir=${CERTS_DIR} --listen-addr=localhost:26257"
    echo "2. Initialize cluster: cockroach init --certs-dir=${CERTS_DIR} --host=localhost:26257"
    echo "3. Connect securely: cockroach sql --certs-dir=${CERTS_DIR} --host=localhost:26257"
    echo "4. Use secure config: cp .env.secure .env"
    echo ""
    echo "Security Notes:"
    echo "- Keep ${CA_KEY_FILE} secure and backed up"
    echo "- Private keys have restrictive permissions (600)"
    echo "- Certificates are valid for 1 year by default"
    echo ""
}

# Main execution function - follows single responsibility
main() {
    echo "================================================"
    echo "CockroachDB Certificate Generation"
    echo "================================================"
    echo ""
    
    # Execute certificate generation steps in order
    check_cockroachdb_installed
    cleanup_existing_certificates
    create_certificate_directories
    generate_ca_certificate
    generate_node_certificate
    generate_client_certificate
    verify_certificates
    set_certificate_permissions
    display_certificate_info
    display_usage_instructions
}

# Execute main function
main "$@"

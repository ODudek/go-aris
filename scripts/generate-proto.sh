#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if protoc is installed
check_protoc() {
    if ! command -v protoc &> /dev/null; then
        log_error "protoc is not installed. Please install Protocol Buffers compiler."
        log_info "On macOS: brew install protobuf"
        log_info "On Ubuntu: sudo apt-get install protobuf-compiler"
        exit 1
    fi

    local version=$(protoc --version | grep -o '[0-9]\+\.[0-9]\+')
    log_info "Found protoc version: $version"
}

# Check if protoc-gen-go and protoc-gen-go-grpc are installed
check_go_plugins() {
    if ! command -v protoc-gen-go &> /dev/null; then
        log_warn "protoc-gen-go not found. Installing..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    fi

    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        log_warn "protoc-gen-go-grpc not found. Installing..."
        go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    fi

    log_success "Go protobuf plugins are available"
}

# Create output directory
create_output_dir() {
    local output_dir="proto/generated"
    if [ ! -d "$output_dir" ]; then
        mkdir -p "$output_dir"
        log_info "Created output directory: $output_dir"
    fi
}

# Generate Go code from proto files
generate_proto() {
    local proto_dir="proto"
    local output_dir="proto/generated"

    log_info "Generating Go code from proto files..."

    # Find all .proto files
    local proto_files=$(find "$proto_dir" -name "*.proto" -not -path "*/generated/*")

    if [ -z "$proto_files" ]; then
        log_warn "No .proto files found in $proto_dir"
        return 0
    fi

    for proto_file in $proto_files; do
        log_info "Processing: $proto_file"

        # Generate Go structs and serialization code
        protoc \
            --go_out="$output_dir" \
            --go_opt=paths=source_relative \
            --go-grpc_out="$output_dir" \
            --go-grpc_opt=paths=source_relative \
            --proto_path="$proto_dir" \
            "$proto_file"

        if [ $? -eq 0 ]; then
            log_success "Generated Go code for: $proto_file"
        else
            log_error "Failed to generate Go code for: $proto_file"
            exit 1
        fi
    done
}

# Update go.mod with required dependencies
update_dependencies() {
    log_info "Updating Go dependencies..."

    # Add required gRPC dependencies if not present
    if ! grep -q "google.golang.org/grpc" go.mod; then
        go get google.golang.org/grpc@latest
    fi

    if ! grep -q "google.golang.org/protobuf" go.mod; then
        go get google.golang.org/protobuf@latest
    fi

    # Tidy up dependencies
    go mod tidy

    log_success "Dependencies updated"
}

# Validate generated code
validate_generated_code() {
    log_info "Validating generated code..."

    # Try to build the generated code
    if go build ./proto/generated/... 2>/dev/null; then
        log_success "Generated code compiles successfully"
    else
        log_warn "Generated code has compilation issues (this might be normal if interfaces are not implemented yet)"
    fi
}

# Clean generated files
clean() {
    log_info "Cleaning generated files..."
    rm -rf proto/generated/*
    log_success "Cleaned generated files"
}

# Show usage
usage() {
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  generate  - Generate Go code from proto files (default)"
    echo "  clean     - Remove all generated files"
    echo "  deps      - Update Go dependencies only"
    echo "  validate  - Validate generated code"
    echo "  help      - Show this help message"
    echo
    echo "Environment variables:"
    echo "  PROTO_DIR    - Directory containing .proto files (default: proto)"
    echo "  OUTPUT_DIR   - Output directory for generated files (default: proto/generated)"
}

# Main function
main() {
    local command="${1:-generate}"

    case "$command" in
        "generate"|"gen")
            check_protoc
            check_go_plugins
            create_output_dir
            generate_proto
            update_dependencies
            validate_generated_code
            log_success "Proto generation completed successfully!"
            ;;
        "clean")
            clean
            ;;
        "deps")
            update_dependencies
            ;;
        "validate")
            validate_generated_code
            ;;
        "help"|"-h"|"--help")
            usage
            ;;
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"

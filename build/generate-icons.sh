#!/bin/bash

# Generates multiple icon sizes from a source PNG file for packaging
# Creates icons in sizes: 16, 32, 48, 64, 128, 256, 512 pixels

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SOURCE_ICON="${SCRIPT_DIR}/hashverifier.png"
readonly OUTPUT_DIR="${ICONS_OUTPUT_DIR:-${SCRIPT_DIR}}"
readonly ICON_SIZES=(16 32 48 64 128 256 512)

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

validate_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v convert &> /dev/null; then
        log_error "ImageMagick 'convert' not found. Install imagemagick."
        exit 1
    fi

    if [[ ! -f "${SOURCE_ICON}" ]]; then
        log_error "Source icon not found: ${SOURCE_ICON}"
        exit 1
    fi

    mkdir -p "${OUTPUT_DIR}"
}

generate_icons() {
    log_info "Generating icons from: ${SOURCE_ICON}"
    log_info "Output directory: ${OUTPUT_DIR}"

    for size in "${ICON_SIZES[@]}"; do
        local output_file="${OUTPUT_DIR}/hashverifier-${size}.png"
        log_info "Generating ${size}x${size}: ${output_file}"

        convert "${SOURCE_ICON}" \
            -resize "${size}x${size}" \
            -quality 100 \
            "${output_file}"
    done

    log_success "Generated ${#ICON_SIZES[@]} icon sizes"
}

main() {
    validate_prerequisites
    generate_icons
}

main "$@"

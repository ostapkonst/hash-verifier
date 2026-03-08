#!/bin/bash

# Creates an RPM package for Red Hat/Fedora systems
# Builds a complete RPM package with all dependencies and desktop integration

set -euo pipefail

readonly PACKAGE_NAME="hashverifier"
readonly PACKAGE_LICENSE="MIT"
readonly PACKAGE_SUMMARY="Cross-platform checksum generation and verification tool"
readonly PACKAGE_URL="https://github.com/ostapkonst/hash-verifier"
readonly PACKAGE_ARCH="${PACKAGE_ARCH:-x86_64}"
readonly PACKAGE_RELEASE="1"
readonly PACKAGE_MAINTAINER="ostapkonst"
readonly PACKAGE_GROUP="Applications/System"

readonly PACKAGE_DEPENDS=(
    "gtk3 >= 3.24"
    "glibc >= 2.34"
)

readonly BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly BUILD_DIR="${BASE_DIR}/build"

# Маппинг архитектуры RPM на архитектуру Docker-сборки
case "${PACKAGE_ARCH}" in
    x86_64)  BUILD_ARCH="amd64" ;;
    aarch64) BUILD_ARCH="arm64" ;;
    *)       BUILD_ARCH="${PACKAGE_ARCH}" ;;
esac

readonly DIST_DIR="${BASE_DIR}/dist/linux-${BUILD_ARCH}"
readonly WORK_DIR="${BASE_DIR}/.pkg-build"
readonly OUT_DIR="${WORK_DIR}/package"
readonly ICONS_DIR="${WORK_DIR}/icons"

RPM_VERSION="${VERSION#v}"
readonly RPM_VERSION="${RPM_VERSION//-/\~}"

readonly SOURCE_BINARY="${DIST_DIR}/${PACKAGE_NAME}"
readonly SOURCE_DESKTOP="${BUILD_DIR}/${PACKAGE_NAME}.desktop"
readonly SOURCE_MIME="${BUILD_DIR}/${PACKAGE_NAME}-mime.xml"
readonly SOURCE_ICON="${BUILD_DIR}/${PACKAGE_NAME}.png"
readonly SOURCE_LICENSE="${BASE_DIR}/LICENSE"
readonly SOURCE_THIRD_PARTY="${BASE_DIR}/THIRD_PARTY_NOTICES"

readonly RPM_ROOT="${WORK_DIR}/dist/rpm/${PACKAGE_NAME}-${RPM_VERSION}-${PACKAGE_ARCH}"
readonly RPM_BUILD_DIR="${RPM_ROOT}/BUILD"
readonly RPM_SOURCE_DIR="${RPM_ROOT}/SOURCES"
readonly RPM_SPEC_DIR="${RPM_ROOT}/SPECS"
readonly RPM_BUILDROOT="${RPM_ROOT}/BUILDROOT"

readonly RPM_PACKAGE_NAME="${PACKAGE_NAME}-${RPM_VERSION}-${PACKAGE_RELEASE}.${PACKAGE_ARCH}.rpm"

log_info() {
    echo "[INFO] $*"
}

log_stage() {
    echo ""
    echo "========================================"
    echo "  STAGE: $*"
    echo "========================================"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

log_warn() {
    echo "[WARNING] $*"
}

validate_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v rpmbuild &> /dev/null; then
        log_error "rpmbuild not found. Install rpm-build."
        exit 1
    fi
}

validate_source_files() {
    log_info "Validating source files..."

    local missing_files=()

    if [[ ! -f "${SOURCE_BINARY}" ]]; then
        missing_files+=("${SOURCE_BINARY}")
    fi

    if [[ ! -f "${SOURCE_DESKTOP}" ]]; then
        missing_files+=("${SOURCE_DESKTOP}")
    fi

    if [[ ! -f "${SOURCE_MIME}" ]]; then
        missing_files+=("${SOURCE_MIME}")
    fi

    if [[ ! -f "${SOURCE_ICON}" ]]; then
        missing_files+=("${SOURCE_ICON}")
    fi

    if [[ ! -f "${SOURCE_LICENSE}" ]]; then
        missing_files+=("${SOURCE_LICENSE}")
    fi

    if [[ ! -f "${SOURCE_THIRD_PARTY}" ]]; then
        missing_files+=("${SOURCE_THIRD_PARTY}")
    fi

    if [[ ${#missing_files[@]} -gt 0 ]]; then
        log_error "Missing required files:"
        for file in "${missing_files[@]}"; do
            log_error "  - ${file}"
        done
        exit 1
    fi

    if [[ -f "${SOURCE_BINARY}" && ! -x "${SOURCE_BINARY}" ]]; then
        log_warn "Binary is not executable, fixing permissions..."
        chmod +x "${SOURCE_BINARY}"
    fi
}

cleanup() {
    local exit_code=$?
    if [[ ${exit_code} -ne 0 ]]; then
        log_error "Error occurred. Temporary files preserved in: ${WORK_DIR}"
    fi
    return ${exit_code}
}

prepare_version() {
    log_info "Processing version: ${RPM_VERSION}"
}

prepare_directories() {
    log_info "Creating directory structure..."

    rm -rf "${RPM_ROOT}"

    mkdir -p "${RPM_ROOT}"
    mkdir -p "${RPM_BUILD_DIR}"
    mkdir -p "${RPM_SOURCE_DIR}"
    mkdir -p "${RPM_SPEC_DIR}"
    mkdir -p "${RPM_BUILDROOT}"
    mkdir -p "${OUT_DIR}"
}

generate_icons() {
    log_info "Generating icon sizes..."

    local generate_script="${BUILD_DIR}/generate-icons.sh"

    mkdir -p "${ICONS_DIR}"

    export ICONS_OUTPUT_DIR="${ICONS_DIR}"

    if [[ -f "${generate_script}" ]]; then
        bash "${generate_script}"
    else
        log_error "Icon generation script not found: ${generate_script}"
        exit 1
    fi
}

prepare_source_tarball() {
    log_info "Preparing source tarball..."

    local source_dir="${RPM_SOURCE_DIR}/${PACKAGE_NAME}-${RPM_VERSION}"

    mkdir -p "${source_dir}"

    cp "${SOURCE_BINARY}" "${source_dir}/"
    cp "${SOURCE_DESKTOP}" "${source_dir}/"
    cp "${SOURCE_MIME}" "${source_dir}/"
    cp "${SOURCE_LICENSE}" "${source_dir}/"
    cp "${SOURCE_THIRD_PARTY}" "${source_dir}/"
    cp "${SOURCE_ICON}" "${source_dir}/"

    for size in 16 32 48 64 128 256 512; do
        local icon_file="${ICONS_DIR}/${PACKAGE_NAME}-${size}.png"
        cp "${icon_file}" "${source_dir}/"
    done

    tar -czf "${RPM_SOURCE_DIR}/${PACKAGE_NAME}-${RPM_VERSION}.tar.gz" \
        -C "${RPM_SOURCE_DIR}" "${PACKAGE_NAME}-${RPM_VERSION}"

    rm -rf "${source_dir}"
}

create_spec_file() {
    log_info "Creating spec file..."

    local dependencies

    dependencies=$(printf "Requires:       %s\n" "${PACKAGE_DEPENDS[@]}")

    cat > "${RPM_SPEC_DIR}/${PACKAGE_NAME}.spec" << EOF
Name:           ${PACKAGE_NAME}
Version:        ${RPM_VERSION}
Release:        ${PACKAGE_RELEASE}%{?dist}
Summary:        ${PACKAGE_SUMMARY}
License:        ${PACKAGE_LICENSE}
Group:          ${PACKAGE_GROUP}
URL:            ${PACKAGE_URL}
ExclusiveArch:  ${PACKAGE_ARCH}
${dependencies}
Source0:        %{name}-%{version}.tar.gz

%description
HashVerifier is a cross-platform checksum generation and verification tool
with both CLI and GTK3 graphical interface.

Features:
- Checksum generation for entire directories
- File verification against checksum files
- Support for 11 hash algorithms (CRC32, MD4, MD5, SHA1, SHA256, SHA384, SHA512,
  SHA3-256, SHA3-384, SHA3-512, BLAKE3)
- CLI and GUI interfaces
- File associations for all checksum formats

%prep
%setup -q

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/applications
mkdir -p %{buildroot}/usr/share/mime/packages
mkdir -p %{buildroot}/usr/share/doc/%{name}
mkdir -p %{buildroot}/usr/share/icons/hicolor/16x16/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/32x32/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/48x48/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/64x64/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/128x128/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/256x256/apps
mkdir -p %{buildroot}/usr/share/icons/hicolor/512x512/apps

cp hashverifier %{buildroot}/usr/bin/
chmod +x %{buildroot}/usr/bin/hashverifier
cp hashverifier.desktop %{buildroot}/usr/share/applications/
cp hashverifier-mime.xml %{buildroot}/usr/share/mime/packages/
cp LICENSE %{buildroot}/usr/share/doc/%{name}/
cp THIRD_PARTY_NOTICES %{buildroot}/usr/share/doc/%{name}/
cp hashverifier-16.png %{buildroot}/usr/share/icons/hicolor/16x16/apps/hashverifier.png
cp hashverifier-32.png %{buildroot}/usr/share/icons/hicolor/32x32/apps/hashverifier.png
cp hashverifier-48.png %{buildroot}/usr/share/icons/hicolor/48x48/apps/hashverifier.png
cp hashverifier-64.png %{buildroot}/usr/share/icons/hicolor/64x64/apps/hashverifier.png
cp hashverifier-128.png %{buildroot}/usr/share/icons/hicolor/128x128/apps/hashverifier.png
cp hashverifier-256.png %{buildroot}/usr/share/icons/hicolor/256x256/apps/hashverifier.png
cp hashverifier-512.png %{buildroot}/usr/share/icons/hicolor/512x512/apps/hashverifier.png

%post
update-desktop-database /usr/share/applications > /dev/null 2>&1 || :
update-mime-database /usr/share/mime > /dev/null 2>&1 || :
gtk-update-icon-cache -f /usr/share/icons/hicolor > /dev/null 2>&1 || :

%postun
update-desktop-database /usr/share/applications > /dev/null 2>&1 || :
update-mime-database /usr/share/mime > /dev/null 2>&1 || :
gtk-update-icon-cache -f /usr/share/icons/hicolor > /dev/null 2>&1 || :

%files
/usr/bin/hashverifier
/usr/share/applications/hashverifier.desktop
/usr/share/mime/packages/hashverifier-mime.xml
/usr/share/doc/%{name}/LICENSE
/usr/share/doc/%{name}/THIRD_PARTY_NOTICES
/usr/share/icons/hicolor/16x16/apps/hashverifier.png
/usr/share/icons/hicolor/32x32/apps/hashverifier.png
/usr/share/icons/hicolor/48x48/apps/hashverifier.png
/usr/share/icons/hicolor/64x64/apps/hashverifier.png
/usr/share/icons/hicolor/128x128/apps/hashverifier.png
/usr/share/icons/hicolor/256x256/apps/hashverifier.png
/usr/share/icons/hicolor/512x512/apps/hashverifier.png
EOF
}

build_package() {
    log_info "Building RPM package..."

    rpmbuild --define "_topdir ${RPM_ROOT}" -bb "${RPM_SPEC_DIR}/${PACKAGE_NAME}.spec" 2>&1 | grep -v "^+" || true

    mv "${RPM_ROOT}/RPMS/${PACKAGE_ARCH}/${RPM_PACKAGE_NAME}" "${OUT_DIR}/"
}

show_package_info() {
    local rpm_file="${OUT_DIR}/${RPM_PACKAGE_NAME}"

    log_stage "Package Information"

    log_info "RPM package details:"
    rpm -qip "${rpm_file}"

    echo ""
    log_info "RPM package requires:"
    rpm -qRp "${rpm_file}"

    echo ""
    log_info "RPM package files:"
    rpm -qlp "${rpm_file}"
}

main() {
    trap cleanup EXIT

    log_stage "Initialization"

    prepare_version
    validate_prerequisites
    validate_source_files

    log_stage "Preparation"

    prepare_directories
    generate_icons
    prepare_source_tarball

    log_stage "Package Build"

    create_spec_file
    build_package
    show_package_info

    log_stage "Complete"

    log_success "RPM package created: ${OUT_DIR}/${RPM_PACKAGE_NAME}"
}

main "$@"

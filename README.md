# HashVerifier

A cross-platform checksum generation and verification tool with both CLI and GTK3 graphical interface.

![HashVerifier GUI](docs/screenshots/main-window.png)

## Features

- **Checksum Generation** — Recursively generate checksum files for entire directories
- **File Verification** — Verify files against existing checksum files
- **Multiple Algorithms** — Support for 11 hash algorithms (see below)
- **Dual Interface** — Use via command-line or intuitive GUI
- **Progress Tracking** — Real-time progress for both generation and verification
- **Detailed Logging** — Structured logging with file sizes and statistics
- **Symbolic Link Support** — Follows symbolic links and hard links (Linux, Windows), and junction points (Windows only)
- **UTF-8 Encoding** — All checksum files are saved in UTF-8 encoding

## Supported Platforms

| Operating System | Architecture | Binary | Package |
|------------------|--------------|--------|---------|
| Linux | x86_64 (amd64) | ✅ | DEB, RPM, AppImage |
| Linux | ARM64 (aarch64) | ✅ | DEB, RPM |
| Windows | x86_64 (amd64) | ✅ | ZIP |
| Windows | x86 (i686) | ✅ | ZIP |

**Minimum OS versions:**

- **Linux:** Any modern distribution with glibc (Debian 11+, Ubuntu 20.04+, Fedora 35+, RHEL 8+)
- **Windows:** Windows 7 SP1 and later (both 32-bit and 64-bit)

> **Note for Windows builds:** The Windows binaries run in GUI mode only (no CLI support). This is a limitation of the Windows operating system, which does not allow combining both GUI and CLI modes in a single application.

## Supported Hash Algorithms

| Algorithm | Extension | Format |
|-----------|-----------|--------|
| CRC32 | `.sfv` | `filename hash` |
| MD4 | `.md4` | `hash *filename` |
| MD5 | `.md5` | `hash *filename` |
| SHA1 | `.sha1` | `hash *filename` |
| SHA256 | `.sha256` | `hash *filename` |
| SHA384 | `.sha384` | `hash *filename` |
| SHA512 | `.sha512` | `hash *filename` |
| SHA3-256 | `.sha3-256` | `hash *filename` |
| SHA3-384 | `.sha3-384` | `hash *filename` |
| SHA3-512 | `.sha3-512` | `hash *filename` |
| BLAKE3 | `.blake3` | `hash *filename` |

## Installation

### Prerequisites

**For native build:**

- Go 1.24 or later
- GTK3 development libraries (for GUI support)

**For Windows builds:**

- For native Windows builds, a special Go compiler with Windows 7 support is used: [go-legacy-win7](https://github.com/thongtech/go-legacy-win7)

**For Docker build:**

- Docker
- docker-compose

### Build from Source

```bash
# Clone repository
git clone https://github.com/ostapkonst/hash-verifier.git
cd hash-verifier

# Native build
make build

# Or build using Docker
make linux-amd64      # Linux binary (x86_64)
make linux-arm64      # Linux binary (ARM64/aarch64)
make windows-amd64    # Windows binary (x86_64)
make windows-i686     # Windows binary (x86, 32-bit)
```

## Usage

### GUI Mode (Default)

Launch the graphical interface:

```bash
# Launch GUI
./hashverifier
```

Open a specific directory or checksum file directly:

```bash
# Open directory (Generate tab)
./hashverifier /path/to/directory

# Open checksum file (Verify tab)
./hashverifier /path/to/checksum.sha256
```

### CLI Mode

#### Generate Checksum File

```bash
# Syntax
./hashverifier generate <input_directory> <output_checksum_file>
```

**Examples:**

```bash
# Generate SHA256 checksums
./hashverifier generate ./data ./data.sha256

# Generate MD5 checksums
./hashverifier generate ./photos ./photos.md5

# Generate CRC32 (SFV) checksums
./hashverifier generate ./backup ./backup.sfv
```

#### Verify Files

```bash
# Syntax
./hashverifier verify <checksum_file>
```

**Examples:**

```bash
# Verify files against SHA256 checksums
./hashverifier verify ./data.sha256

# Verify files against MD5 checksums
./hashverifier verify ./archive.md5
```

### Output Format

Checksum files include a header with the generation timestamp. Two formats are used:

| Format | Algorithms | Example |
|--------|------------|---------|
| **Hash-first** | MD4, MD5, SHA1, SHA256, SHA384, SHA512, SHA3-256, SHA3-384, SHA3-512, BLAKE3 | `<hash> *<path>` |
| **Path-first** | CRC32 (`.sfv`) | `<path> <hash>` |

**Example (SHA256):**

```
; Generated at <timestamp> by HashVerifier (<link>)

a1b2c3d4e5f6... *documents/report.pdf
f6e5d4c3b2a1... *documents/notes.txt
```

**Example (CRC32/SFV):**

```
; Generated at <timestamp> by HashVerifier (<link>)

documents/report.pdf a1b2c3d4
documents/notes.txt f6e5d4c3
```

## Verification Results

During file verification, each file is assigned a status:

| Status | Description |
|--------|-------------|
| `MATCHED` | File hash matches the expected value — integrity confirmed |
| `MISMATCH` | File hash differs from expected — file may be corrupted or modified |
| `UNREADABLE` | File could not be read — missing, permission denied, or I/O error |

Results are logged to the console with detailed information including file path, expected/actual hash, and file size.

## Project Structure

```
hash-verifier/
├── src/                  # Go source code
├── build/                # Docker build files (Dockerfile, docker-compose, scripts)
├── .github/workflows/    # GitHub Actions CI/CD
├── dist/                 # Build output (git-ignored)
├── .pkg-build/           # Package build temporary files (git-ignored)
├── docs/                 # Documentation (screenshots, etc.)
├── Makefile              # Build automation
├── .golangci.yml         # Go linter configuration
├── .gitattributes        # Git attributes (line endings, binary files)
├── .gitignore            # Git ignore rules
├── .dockerignore         # Docker build context exclusions
├── LICENSE               # MIT License
├── THIRD_PARTY_NOTICES   # Third-party software notices and licenses
└── README.md             # This file
```

## Building with Docker

The project includes Docker-based cross-compilation for Linux and Windows:

```bash
# Build Linux binaries
make linux-amd64      # x86_64
make linux-arm64      # ARM64/aarch64

# Build Windows binaries
make windows-amd64    # x86_64
make windows-i686     # x86 (32-bit)

# Build Linux packages
make deb-amd64        # DEB package for amd64
make deb-arm64        # DEB package for ARM64
make rpm-amd64        # RPM package for amd64
make rpm-arm64        # RPM package for ARM64
make appimage         # AppImage (amd64 only)
```

### Linux Packages

**DEB (Debian/Ubuntu):**

```bash
# AMD64 (x86_64)
sudo apt install ./hashverifier_1.0.0_amd64.deb

# ARM64 (aarch64)
sudo apt install ./hashverifier_1.0.0_arm64.deb
```

**RPM (Fedora/RHEL):**

```bash
# AMD64 (x86_64)
sudo dnf install ./hashverifier-1.0.0-1.x86_64.rpm

# ARM64 (aarch64)
sudo dnf install ./hashverifier-1.0.0-1.aarch64.rpm
```

**AppImage (Universal Linux, AMD64 only):**

Download the AppImage file and make it executable:

```bash
chmod +x HashVerifier-1.0.0-x86_64.AppImage
./HashVerifier-1.0.0-x86_64.AppImage
```

DEB and RPM packages include:

- Binary in `/usr/bin/hashverifier`
- Automatic dependency installation (GTK3 and related libraries)
- **File associations** for all checksum formats:
  - `.sfv`, `.md4`, `.md5`, `.sha1`
  - `.sha256`, `.sha384`, `.sha512`
  - `.sha3-256`, `.sha3-384`, `.sha3-512`
  - `.blake3`
- Desktop integration (right-click → Open with HashVerifier)

### Windows Binary

For Windows, download and extract the ZIP archive for your architecture:

```bash
# 64-bit Windows (x86_64)
hashverifier-vX.X.X-windows-amd64.zip

# 32-bit Windows (x86)
hashverifier-vX.X.X-windows-i686.zip
```

The ZIP archive contains the executable and all required DLLs.

## Technologies

- **Language:** Go 1.24
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **GUI Toolkit:** [gotk3](https://github.com/gotk3/gotk3) (GTK3 bindings)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **Cryptography:** [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto), [blake3](https://github.com/lukechampine/blake3)

## Related Projects

HashVerifier was inspired by the following Windows-only checksum tools:

- **[HashCheck Shell Extension](https://github.com/gurnec/HashCheck)** — Windows shell extension for checksum generation and verification (originally from code.kliu.org)
- **[HashCheck Fork](https://github.com/idrassi/HashCheck)** — Enhanced fork of HashCheck with additional features and bug fixes

Unlike these Windows-only tools, HashVerifier is cross-platform and works on Linux and Windows.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.

For third-party software notices, see [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES).

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Branch and Pull Request Naming Convention

To maintain a clear project history, please follow these naming conventions:

**Branch names:**
- `feature/<description>` — for new features
- `fix/<description>` — for bug fixes

**Pull Request titles:**
- `FEATURE: <description>` — for new features
- `FIX: <description>` — for bug fixes

**Examples:**

| Type | Branch Name | PR Title |
|------|-------------|----------|
| Feature | `feature/linux-arm64-support` | `FEATURE: Add Linux ARM64 (aarch64) build support` |
| Feature | `feature/dark-mode-theme` | `FEATURE: Implement dark mode theme for GUI` |
| Fix | `fix/windows-gtk3-build-error` | `FIX: Resolve GTK3 build error on Windows with MinGW` |
| Fix | `fix/sha256-verification-crash` | `FIX: Fix crash when verifying corrupted SHA256 files` |

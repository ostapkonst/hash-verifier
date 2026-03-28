# HashVerifier Development Guide

This document contains information for developers and builders of the project.

## Requirements

**For native build:**

- Go 1.24 or later
- GTK3 development libraries (for GUI)

**For Windows builds:**

- For native Windows builds, a special Go compiler with Windows 7 support is used: [go-legacy-win7](https://github.com/thongtech/go-legacy-win7)

**For Docker build:**

- Docker
- docker-compose

## Build from Source

```bash
# Clone repository
git clone https://github.com/ostapkonst/HashVerifier.git
cd HashVerifier

# Native build
make build

# Or build using Docker
make linux-amd64      # Linux binary (x86_64)
make linux-arm64      # Linux binary (ARM64/aarch64)
make windows-amd64    # Windows binary (x86_64)
make windows-i686     # Windows binary (x86, 32-bit)
```

## Building with Docker

The project includes Docker-based cross-compilation:

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

# Build Linux (Universal) packages
make appimage-amd64   # AppImage for amd64
make appimage-arm64   # AppImage for ARM64
```

## Project Structure

```
HashVerifier/
├── .github/workflows/    # GitHub Actions CI/CD
├── .pkg-build/           # Package build temporary files (git-ignored)
├── src/                  # Go source code
├── build/                # Docker build files (Dockerfile, docker-compose, scripts)
├── flatpak/              # Required to publish an application on FlatHub
├── dist/                 # Build output (git-ignored)
├── docs/                 # Documentation
├── .golangci.yml         # Go linter configuration
├── .gitattributes        # Git attributes (line endings, binary files)
├── .gitignore            # Git ignore rules
├── .dockerignore         # Docker build context exclusions
├── LICENSE               # MIT License
├── THIRD_PARTY_NOTICES   # Third-party software notices
├── README.md             # Main documentation
└── Makefile              # Build automation
```

## Technologies

- **Language:** Go 1.24
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **GUI Toolkit:** [gotk3](https://github.com/gotk3/gotk3) (GTK3 bindings)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **Cryptography:** [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto), [blake3](https://github.com/lukechampine/blake3)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

# HashVerifier

A cross-platform checksum generation and verification tool with both CLI and GTK3 graphical interface.

![HashVerifier GUI](docs/screenshots/main-window.png)

## Features

- **Checksum Generation** — Recursively generate checksum files for entire directories
- **File Verification** — Verify files against existing checksum files
- **Multiple Algorithms** — Support for 11 hash algorithms
- **Dual Interface** — Use via command-line or intuitive GUI
- **Progress Tracking** — Real-time progress for both generation and verification
- **Speed Tracking** — Live hashing speed display
- **Symbolic Link Support** — Follows symbolic links, hard links, and junction points
- **UTF-8 Encoding** — All checksum files are saved in UTF-8 encoding

## Supported Platforms

| Operating System | Architecture | Binary | Package |
|------------------|--------------|--------|---------|
| Linux | x86_64 (amd64) | ✅ | DEB, RPM, AppImage |
| Linux | ARM64 (aarch64) | ✅ | DEB, RPM |
| Windows | x86_64 (amd64) | ✅ | ZIP |
| Windows | x86 (i686) | ✅ | ZIP |

**Minimum OS versions:**

- **Linux:** Debian 11+, Ubuntu 20.04+, Fedora 35+, RHEL 8+
- **Windows:** Windows 7 SP1 and later (32-bit and 64-bit)

> **Note for Windows:** Windows binaries run in GUI mode only (no CLI support).

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

### Linux

**DEB (Debian/Ubuntu):**

```bash
sudo apt install ./hashverifier_1.0.0_amd64.deb
```

**RPM (Fedora/RHEL):**

```bash
sudo dnf install ./hashverifier-1.0.0-1.x86_64.rpm
```

**AppImage (Universal Linux):**

```bash
chmod +x HashVerifier-1.0.0-x86_64.AppImage
./HashVerifier-1.0.0-x86_64.AppImage
```

### Windows

Download and extract the ZIP archive for your architecture:

- `hashverifier-vX.X.X-windows-amd64.zip` (64-bit)
- `hashverifier-vX.X.X-windows-i686.zip` (32-bit)

## Usage

### GUI Mode (Default)

```bash
# Launch GUI
./hashverifier

# Open directory (Generate tab)
./hashverifier /path/to/directory

# Open checksum file (Verify tab)
./hashverifier /path/to/checksum.sha256
```

### CLI Mode

**Print version:**

```bash
./hashverifier --version
./hashverifier -v
```

**Generate checksums:**

```bash
./hashverifier generate ./data ./data.sha256
./hashverifier generate ./photos ./photos.md5
```

**Verify files:**

```bash
./hashverifier verify ./data.sha256
./hashverifier verify ./archive.md5
```

### Output Format

**SHA256 example:**

```
; Generated at <timestamp> by HashVerifier <version>

a1b2c3d4e5f6... *documents/report.pdf
f6e5d4c3b2a1... *documents/notes.txt
```

**CRC32/SFV example:**

```
; Generated at <timestamp> by HashVerifier <version>

documents/report.pdf a1b2c3d4
documents/notes.txt f6e5d4c3
```

### Verification Results

| Status | Description |
|--------|-------------|
| `MATCHED` | File hash matches — integrity confirmed |
| `MISMATCH` | File hash differs — file may be corrupted |
| `UNREADABLE` | File could not be read — missing or permission denied |

## Build from Source

See [Development Guide](docs/DEVELOPMENT.md) for build instructions and contribution guidelines.

## Related Projects

HashVerifier was inspired by:

- [HashCheck Shell Extension](https://github.com/gurnec/HashCheck)
- [HashCheck Fork](https://github.com/idrassi/HashCheck)

Unlike these Windows-only tools, HashVerifier is cross-platform.

## License

MIT License — see [LICENSE](LICENSE). Third-party notices in [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES).

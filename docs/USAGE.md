# HashVerifier Usage Guide

## GUI Mode (Default)

```bash
# Launch GUI
./hashverifier

# Open directory (Generate tab)
./hashverifier /path/to/directory

# Open checksum file (Verify tab)
./hashverifier /path/to/checksum.sha256
```

## CLI Mode

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

## Configuration

See [Configuration Guide](CONFIGURATION.md) for detailed settings documentation.

**Quick commands:**

```bash
./hashverifier config        # View settings
./hashverifier config edit   # Edit settings
./hashverifier config reset  # Reset to defaults
```

## Output Format

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

**Footer with statistics (appended to all checksum files):**

```
; Statistics:
;   Status: success
;   Processed: 2
```

**Status values:**

| Status | Description |
|--------|-------------|
| `success` | All files were hashed successfully |
| `completed with errors` | Some files could not be hashed (e.g., permission denied) |
| `cancelled` | Operation was cancelled by the user |

## Verification Results

| Status | Description |
|--------|-------------|
| `MATCHED` | File hash matches — integrity confirmed |
| `MISMATCH` | File hash differs — file may be corrupted |
| `UNREADABLE` | File could not be read — missing or permission denied |

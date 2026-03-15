# HashVerifier Configuration

HashVerifier stores user preferences in a YAML settings file. You can view and edit settings via the command line or by directly editing the settings file.

## Settings Location

| Platform | Path |
|----------|------|
| Linux | `~/.config/hashverifier/settings.yaml` |
| Windows | `%APPDATA%\hashverifier\settings.yaml` |

## CLI Commands

**View settings:**

```bash
./hashverifier config
./hashverifier config show
```

**Edit settings:**

```bash
./hashverifier config edit
```

Opens the settings file in your default text editor (`$VISUAL` or `$EDITOR`).

**Reset settings:**

```bash
./hashverifier config reset
```

## Available Settings

### Window Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `window.tab_order` | `generate, verify` | Order of tabs in main window |
| `window.current_page` | `0` | Currently active tab |

### Generate Tab Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `generate.follow_symbolic_links` | `true` | Follow symbolic links when scanning directories |
| `generate.sort_paths` | `true` | Sort paths before hashing |
| `generate.algorithm` | `.md5` | Default hash algorithm |
| `generate.column_order` | `path, size, hash, note` | Order of columns in Generate tab |
| `generate.sort_column` | `path` | Column to sort by in Generate tab |
| `generate.sort_order` | `asc` | Sort order in Generate tab (asc/desc) |

### Verify Tab Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `verify.verify_on_open` | `true` | Auto-start verification when opening checksum file |
| `verify.column_order` | `status, path, size, hash, expected_hash, note` | Order of columns in Verify tab |
| `verify.sort_column` | `status` | Column to sort by in Verify tab |
| `verify.sort_order` | `desc` | Sort order in Verify tab (asc/desc) |

### Flatpak Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `flatpak.suppress_sandbox_warning` | `false` | Suppress the Flatpak sandbox warning dialog on startup |

> **Note:** Flatpak settings only apply when running the application as a Flatpak package.

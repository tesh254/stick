# Getting Started with Stick

Stick is a powerful CLI tool that enhances your Git workflow. This guide will help you get started with installation, basic usage, and maintenance of the tool.

## Platform Support

**Currently supported platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)

**Windows is not currently supported** for the main stick binary. However, you can use Windows Subsystem for Linux (WSL) to run stick.

## Installation

### Method 1: Using the Install Script (Recommended)

The install script will automatically download and install stick along with its required dependency, Turso CLI, which provides local database functionality for storing embeddings and other data:

```bash
curl -fsSL https://raw.githubusercontent.com/tesh254/stick/main/install.sh | bash
```

Or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/tesh254/stick/main/install.sh | bash
```

> **Note**: The install script will automatically install Turso CLI for local database functionality as a requirement. You'll see installation progress messages during the process.

### Method 2: Using Go

```bash
go install github.com/tesh254/stick@latest
```

> **Note**: If using this method, you'll also need to install Turso CLI separately:
```bash
curl --proto '=https' --tlsv1.2 -LsSf https://github.com/tursodatabase/turso/releases/latest/download/turso_cli-installer.sh | sh
```

### Method 3: Manual Installation

1. Visit the [releases page](https://github.com/tesh254/stick/releases)
2. Download the appropriate binary for your OS and architecture:
   - Linux: `stick-linux-amd64`, `stick-linux-arm64`
   - macOS: `stick-darwin-amd64`, `stick-darwin-arm64`
3. Make it executable: `chmod +x stick-*`
4. Move it to a directory in your PATH: `sudo mv stick-* /usr/local/bin/stick`
5. Install Turso CLI separately: `curl --proto '=https' --tlsv1.2 -LsSf https://github.com/tursodatabase/turso/releases/latest/download/turso_cli-installer.sh | sh`

> **Note**: The tool is also available with the short alias `stk`, so you can use either `stick` or `stk` for all commands.

## File Locations

During installation, stick creates several important files and directories:

- **Binaries**: `stick` and `stk` are installed in your PATH (usually `/usr/local/bin` or `~/.local/bin`)
- **Configuration directory**: `~/.stick/config` - stores the database encryption key and other config
- **Database file**: `~/.stick/data/stick.db` - persistent encrypted database for embeddings and data
- **Config files**: `~/.stick.yaml`, `~/.stick.yml`, or `~/.stick.json` in your home directory

## Basic Usage

### Local Database (Turso)

Stick uses Turso CLI for local database functionality, which enables it to store embeddings and other data locally. During installation, an encrypted database is automatically set up using a randomly generated encryption key.

### Database Encryption

The local database is encrypted using a 256-bit encryption key that is automatically generated during installation using `openssl rand -hex 32`. The database uses AEGIS-256 encryption algorithm via tursodb. The persistent encrypted database file is created at `~/.stick/data/stick.db` and will be used by the application for storing embeddings and other data. The encryption key is stored securely in your configuration directory (`~/.stick/config/db_encryption_key`) with restricted permissions (600) to ensure your data remains private and secure.

### Version Information

Check the current version of stick:

```bash
stick version
# or with the alias
stk version
```

For more detailed information:

```bash
stick version --json    # Output in JSON format
stick version --short   # Show only the version number
stick buildinfo         # Show comprehensive build information
```

### Self-Update

Keep your stick installation up-to-date:

```bash
stick update
# or
stk update
```

This command will:
- Check for the latest release on GitHub
- Compare with your current version
- Prompt for confirmation before updating
- Download and install the new version

## Available Commands

- `stick` or `stk` - Main command
- `stick version` or `stk version` - Show version information
- `stick version --json` - Show version in JSON format
- `stick version --short` - Show only the version number
- `stick buildinfo` - Show comprehensive build information
- `stick update` or `stk update` - Update to the latest version
- `stick help` or `stk help` - Show help information

## Configuration

Stick supports configuration via YAML, JSON, or other formats. The configuration file is typically located at:

- `~/.stick.yaml`
- `$XDG_CONFIG_HOME/stick/config.yaml` or `~/.config/stick/config.yaml`

## Uninstalling

To completely remove stick and all related files:

```bash
curl -fsSL https://raw.githubusercontent.com/tesh254/stick/main/uninstall.sh | bash
```

Or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/tesh254/stick/main/uninstall.sh | bash
```

The uninstall script will:
- Find all stick and stk binaries on your system
- Remove configuration files
- Prompt for confirmation before removal
- Handle both user and system-wide installations

## Troubleshooting

### Command not found
If you get a "command not found" error after installation, ensure that the installation directory is in your PATH.

### Permission errors during installation/update
If you encounter permission errors, you may need to run the install script with appropriate privileges, or install to a user directory.

### Update fails
If the update command fails, you can always manually download the latest release from the [GitHub releases page](https://github.com/tesh254/stick/releases).

## Need Help?

- Check the help: `stick help` or `stick --help`
- Visit the [GitHub repository](https://github.com/tesh254/stick) for more information
- Create an issue if you encounter any problems
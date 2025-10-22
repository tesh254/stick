#!/bin/bash

# uninstall.sh - Uninstall stick CLI tool and all related files
# Usage: curl -fsSL https://raw.githubusercontent.com/tesh254/stick/main/uninstall.sh | bash
# Or: wget -qO- https://raw.githubusercontent.com/tesh254/stick/main/uninstall.sh | bash

set -e

# Function to print messages
print_msg() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

print_warn() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1" >&2
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

# Check if running as root (not always necessary but good to know)
if [ "$EUID" -eq 0 ]; then
    print_warn "Running as root - this may be necessary to remove system-wide installations"
fi

# Find and list stick installations
find_stick_installations() {
    print_msg "Searching for stick installations..."

    # Common installation directories
    INSTALL_DIRS=("/usr/local/bin" "/usr/bin" "/opt/bin")
    
    # Add user-specific directories
    if [ -n "$HOME" ]; then
        INSTALL_DIRS+=("$HOME/bin" "$HOME/.local/bin")
    fi

    FOUND_BINARIES=()
    for dir in "${INSTALL_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            # Check for stick binary
            if [ -f "$dir/stick" ] || [ -L "$dir/stick" ]; then
                if file "$dir/stick" 2>/dev/null | grep -q "stick\|stick"; then
                    FOUND_BINARIES+=("$dir/stick")
                fi
            fi
            
            # Check for stk binary (alias)
            if [ -f "$dir/stk" ] || [ -L "$dir/stk" ]; then
                if file "$dir/stk" 2>/dev/null | grep -q "stick\|stick"; then
                    FOUND_BINARIES+=("$dir/stk")
                fi
            fi
        fi
    done

    if [ ${#FOUND_BINARIES[@]} -eq 0 ]; then
        print_msg "No stick installations found"
        return 1
    else
        print_msg "Found the following stick installations:"
        for binary in "${FOUND_BINARIES[@]}"; do
            echo "  - $binary"
        done
        return 0
    fi
}

# Find and list configuration files
find_config_files() {
    print_msg "Searching for stick configuration files..."

    CONFIG_FILES=()
    
    # Check user's home directory for config files
    if [ -n "$HOME" ]; then
        for conf_file in "$HOME/.stick.yaml" "$HOME/.stick.yml" "$HOME/.stick.json"; do
            if [ -f "$conf_file" ]; then
                CONFIG_FILES+=("$conf_file")
                echo "  - $conf_file"
            fi
        done
        
        # Check for config in XDG config directory
        if [ -n "$XDG_CONFIG_HOME" ] && [ -f "$XDG_CONFIG_HOME/stick/config.yaml" ]; then
            CONFIG_FILES+=("$XDG_CONFIG_HOME/stick/config.yaml")
            echo "  - $XDG_CONFIG_HOME/stick/config.yaml"
        elif [ -n "$HOME" ] && [ -f "$HOME/.config/stick/config.yaml" ]; then
            CONFIG_FILES+=("$HOME/.config/stick/config.yaml")
            echo "  - $HOME/.config/stick/config.yaml"
        fi
        
        # Check for the new config location
        if [ -f "$HOME/.stick/config/db_encryption_key" ]; then
            CONFIG_FILES+=("$HOME/.stick/config/db_encryption_key")
            echo "  - $HOME/.stick/config/db_encryption_key"
        fi
    fi

    if [ ${#CONFIG_FILES[@]} -eq 0 ]; then
        print_msg "No configuration files found"
    else
        print_msg "Found ${#CONFIG_FILES[@]} configuration file(s)"
    fi
}

# Remove stick binaries
remove_binaries() {
    if [ ${#FOUND_BINARIES[@]} -eq 0 ]; then
        print_msg "No binaries to remove"
        return 0
    fi

    print_msg "Removing stick binaries..."
    for binary in "${FOUND_BINARIES[@]}"; do
        print_msg "Removing $binary"
        # Use sudo if needed (for system directories)
        if [[ "$binary" == /usr/* ]] || [[ "$binary" == /opt/* ]]; then
            sudo rm -f "$binary"
        else
            rm -f "$binary"
        fi
    done
    
    print_success "Binaries removed"
}

# Remove configuration files
remove_config_files() {
    if [ ${#CONFIG_FILES[@]} -eq 0 ]; then
        print_msg "No configuration files to remove"
        return 0
    fi

    print_msg "Removing configuration files..."
    for config_file in "${CONFIG_FILES[@]}"; do
        print_msg "Removing $config_file"
        rm -f "$config_file"
    done
    
    # Try to remove empty config directory
    if [ -n "$XDG_CONFIG_HOME" ] && [ -d "$XDG_CONFIG_HOME/stick" ]; then
        if [ -z "$(ls -A "$XDG_CONFIG_HOME/stick")" ]; then
            rmdir "$XDG_CONFIG_HOME/stick" 2>/dev/null || true
            print_msg "Removed empty config directory: $XDG_CONFIG_HOME/stick"
        fi
    elif [ -n "$HOME" ] && [ -d "$HOME/.config/stick" ]; then
        if [ -z "$(ls -A "$HOME/.config/stick")" ]; then
            rmdir "$HOME/.config/stick" 2>/dev/null || true
            print_msg "Removed empty config directory: $HOME/.config/stick"
        fi
    fi
    
    # Also try to remove the new config directory
    if [ -n "$HOME" ] && [ -d "$HOME/.stick/config" ]; then
        if [ -z "$(ls -A "$HOME/.stick/config")" ]; then
            rmdir "$HOME/.stick/config" 2>/dev/null || true
            print_msg "Removed empty config directory: $HOME/.stick/config"
        fi
    fi
    
    print_success "Configuration files removed"
}

# Find and list database files
find_database_files() {
    print_msg "Searching for stick database files..."

    DATABASE_FILES=()
    
    # Check for the encrypted database file
    if [ -n "$HOME" ]; then
        DB_PATH="$HOME/.stick/data/stick.db"
        if [ -f "$DB_PATH" ]; then
            DATABASE_FILES+=("$DB_PATH")
            echo "  - $DB_PATH"
        fi
        
        # Check for any other related database files (including temp or journal files)
        for db_file in "$HOME/.stick/data/stick.db-"*; do
            if [ -f "$db_file" ]; then
                DATABASE_FILES+=("$db_file")
                echo "  - $db_file"
            fi
        done
    fi

    if [ ${#DATABASE_FILES[@]} -eq 0 ]; then
        print_msg "No database files found"
    else
        print_msg "Found ${#DATABASE_FILES[@]} database file(s)"
    fi
}

# Remove database files
remove_database_files() {
    if [ ${#DATABASE_FILES[@]} -eq 0 ]; then
        print_msg "No database files to remove"
        return 0
    fi

    print_msg "Removing database files..."
    for db_file in "${DATABASE_FILES[@]}"; do
        print_msg "Removing $db_file"
        rm -f "$db_file"
    done
    
    # Try to remove empty data directory
    if [ -n "$HOME" ] && [ -d "$HOME/.stick/data" ]; then
        if [ -z "$(ls -A "$HOME/.stick/data")" ]; then
            rmdir "$HOME/.stick/data" 2>/dev/null || true
            print_msg "Removed empty data directory: $HOME/.stick/data"
        fi
    fi
    
    print_success "Database files removed"
}

# Main execution
main() {
    print_msg "Uninstalling stick CLI tool and related files..."
    print_warn "This will completely remove stick and all its configurations, including the encrypted database."
    
    # Confirm with user
    echo -n "Do you want to continue? [y/N]: "
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_msg "Uninstall cancelled."
        exit 0
    fi

    # Find installations
    find_stick_installations
    find_config_files
    find_database_files

    # Remove binaries
    remove_binaries

    # Remove configuration files
    remove_config_files

    # Remove database files
    remove_database_files

    # Try to remove the main .stick directory if it's empty
    if [ -n "$HOME" ] && [ -d "$HOME/.stick" ]; then
        if [ -z "$(ls -A "$HOME/.stick")" ]; then
            rmdir "$HOME/.stick" 2>/dev/null || true
            print_msg "Removed empty .stick directory: $HOME/.stick"
        fi
    fi

    print_success "stick has been completely uninstalled!"
    print_msg "If you installed stick via package manager, please use that to uninstall instead."
    print_msg "To verify complete removal, you can check your PATH directories manually."
}

# Run main function
main "$@"
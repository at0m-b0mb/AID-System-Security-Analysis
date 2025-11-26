#!/bin/bash
# setup.sh - Setup script for Linux deployment

echo "Setting up AID System..."

# Make binary executable
chmod +x aid-system-linux

# Create necessary directories
mkdir -p insulinlogs
mkdir -p Login

# Initialize database if it doesn't exist
if [ ! -f "Login/aid.db" ]; then
    echo "Initializing database..."
    # Prefer using the built binary to initialize the DB. This avoids requiring sqlite3 CLI.
    if [ -x "./aid-system-linux" ]; then
        echo "Using ./aid-system-linux --init to create the database"
        ./aid-system-linux --init
    else
        echo "aid-system-linux binary not found or not executable. Falling back to sqlite3 CLI if available."
        if command -v sqlite3 >/dev/null 2>&1; then
            sqlite3 Login/aid.db < Login/queries.sql
        else
            echo "ERROR: neither ./aid-system-linux nor sqlite3 CLI are available. Please build the binary (see README) or install sqlite3."
            exit 1
        fi
    fi
fi

# Create sample glucose data if not exists
if [ ! -f "glucose_readings_PA1993.csv" ]; then
    echo "Creating sample glucose data..."
    echo "11/5/2025 8:00,95
11/5/2025 8:05,115
11/5/2025 8:10,65
11/5/2025 8:15,197
11/5/2025 8:20,110" > glucose_readings_PA1993.csv
fi

if [ ! -f "glucose_readings_PA2000.csv" ]; then
    echo "11/9/2025 8:00,102
11/9/2025 8:05,118
11/9/2025 8:10,95" > glucose_readings_PA2000.csv
fi

echo "Setup complete!"
echo "Run the application with: ./aid-system-linux"
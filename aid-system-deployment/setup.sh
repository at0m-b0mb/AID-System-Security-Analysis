#!/bin/bash
# setup.sh - Setup script for Linux deployment
# Team Logan - AID System Security Analysis Phase II

set -e

echo "========================================"
echo "   AID System - Setup Script"
echo "   Team Logan - Phase II"
echo "========================================"
echo ""

# Check if Go is installed (for building if needed)
check_go() {
    if command -v go >/dev/null 2>&1; then
        echo "[OK] Go is installed: $(go version)"
        return 0
    else
        echo "[WARNING] Go is not installed - cannot build from source"
        return 1
    fi
}

# Build binary if not exists
build_binary() {
    if [ ! -f "aid-system-linux" ]; then
        echo "[INFO] Binary not found, building from source..."
        if check_go; then
            go build -o aid-system-linux ./cmd/main.go
            echo "[OK] Binary built successfully"
        else
            echo "[ERROR] Cannot build binary - Go is not installed"
            echo "Please install Go from https://golang.org/dl/"
            exit 1
        fi
    else
        echo "[OK] Binary exists: aid-system-linux"
    fi
}

echo "Step 1: Checking and building binary..."
build_binary

# Make binary executable
chmod +x aid-system-linux
echo "[OK] Binary is executable"

# Make exploit.sh executable
if [ -f "exploit.sh" ]; then
    chmod +x exploit.sh
    echo "[OK] exploit.sh is executable"
fi

echo ""
echo "Step 2: Creating required directories..."

# Create necessary directories
mkdir -p insulinlogs
mkdir -p Login
mkdir -p glucose
mkdir -p alerts

echo "[OK] Created directories: insulinlogs, Login, glucose, alerts"

echo ""
echo "Step 3: Initializing database..."

# Initialize database if it doesn't exist
if [ ! -f "Login/aid.db" ]; then
    echo "[INFO] Database not found, initializing..."
    if [ -x "./aid-system-linux" ]; then
        ./aid-system-linux --init
        echo "[OK] Database schema created"
    else
        echo "[ERROR] Binary not executable"
        exit 1
    fi
else
    echo "[OK] Database already exists: Login/aid.db"
fi

# Check if database has users, if not populate from queries.sql
USER_COUNT=$(sqlite3 Login/aid.db "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "0")
if [ "$USER_COUNT" = "0" ]; then
    echo "[INFO] Database is empty, loading seed data..."
    if [ -f "Login/queries.sql" ] && command -v sqlite3 >/dev/null 2>&1; then
        sqlite3 Login/aid.db < Login/queries.sql
        echo "[OK] Seed data loaded from Login/queries.sql"
    else
        echo "[WARNING] Cannot load seed data - sqlite3 not available or queries.sql missing"
    fi
else
    echo "[OK] Database has $USER_COUNT user(s)"
fi

echo ""
echo "Step 4: Creating sample glucose data..."

# Create sample glucose data in the glucose directory
if [ ! -f "glucose/glucose_readings_PA1993.csv" ]; then
    echo "11/5/2025 8:00,95
11/5/2025 8:05,115
11/5/2025 8:10,65
11/5/2025 8:15,197
11/5/2025 8:20,110" > glucose/glucose_readings_PA1993.csv
    echo "[OK] Created glucose/glucose_readings_PA1993.csv"
else
    echo "[OK] glucose/glucose_readings_PA1993.csv already exists"
fi

if [ ! -f "glucose/glucose_readings_PA2000.csv" ]; then
    echo "11/9/2025 8:00,102
11/9/2025 8:05,118
11/9/2025 8:10,95" > glucose/glucose_readings_PA2000.csv
    echo "[OK] Created glucose/glucose_readings_PA2000.csv"
else
    echo "[OK] glucose/glucose_readings_PA2000.csv already exists"
fi

# Also create in root for backward compatibility
if [ ! -f "glucose_readings_PA1993.csv" ]; then
    cp glucose/glucose_readings_PA1993.csv glucose_readings_PA1993.csv
fi
if [ ! -f "glucose_readings_PA2000.csv" ]; then
    cp glucose/glucose_readings_PA2000.csv glucose_readings_PA2000.csv
fi

echo ""
echo "========================================"
echo "   Setup Complete!"
echo "========================================"
echo ""
echo "Available commands:"
echo "  ./aid-system-linux           - Run the AID System (normal mode)"
echo "  ./aid-system-linux --debug   - Run with debug mode (exposes vulnerabilities)"
echo "  ./aid-system-linux --nolog   - Run with logging disabled"
echo "  ./exploit.sh --all           - Run all vulnerability exploits"
echo "  ./exploit.sh --demo          - Interactive exploit demo"
echo ""
echo "Test credentials (from Login/queries.sql):"
echo "  Patient:    PA1993 / PIN: Passw0rd!"
echo "  Clinician:  DR095  / PIN: Cl1n1c1an!"
echo "  Caretaker:  CR055  / PIN: Passw0rd!"
echo ""
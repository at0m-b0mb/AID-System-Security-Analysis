# Artificial Insulin Delivery (AID) System - User Manual

## Table of Contents
1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Installation Guide](#installation-guide)
4. [User Roles](#user-roles)
5. [Getting Started](#getting-started)
6. [Features](#features)
7. [Patient Dashboard](#patient-dashboard)
8. [Clinician Dashboard](#clinician-dashboard)
9. [Caretaker Dashboard](#caretaker-dashboard)
10. [Safety Features](#safety-features)
11. [Troubleshooting](#troubleshooting)

---

## Overview

The **Artificial Insulin Delivery (AID) System** is a command-line medical application designed for diabetes care management. It enables patients, clinicians, and caretakers to collaborate on insulin delivery management, glucose monitoring, and health tracking.

### Key Capabilities
- 🩺 **Patient Management** – Register, track, and manage multiple patients
- 💉 **Insulin Management** – Request and approve basal and bolus insulin doses
- 📊 **Glucose Monitoring** – Real-time glucose tracking with configurable alerts
- ⛔ **Safety Features** – Automatic insulin suspension for critical hypoglycemia
- 🔐 **Security** – Role-based access control with PIN authentication
- 📝 **Audit Trail** – Comprehensive action logging for compliance

---

## System Requirements

### Prerequisites
- **Operating System:** Linux, macOS, or Windows with WSL
- **Go:** Version 1.20 or higher (for building from source)
- **SQLite3:** For database operations (optional, for manual queries)
- **Terminal:** Command-line interface with UTF-8 support

### Supported Platforms
- Linux (x86_64)
- macOS (Intel and Apple Silicon)
- Windows Subsystem for Linux (WSL2)

---

## Installation Guide

### Step 1: Download/Clone the Repository

```bash
# Navigate to the project directory
cd aid-system-deployment
```

### Step 2: Make Setup Script Executable

```bash
chmod +x setup.sh
```

### Step 3: Run Setup Script

```bash
./setup.sh
```

The setup script will:
- Install required Go dependencies
- Build the binary from source (if Go is installed)
- Create required directories (glucose, alerts, insulinlogs, Login)
- Initialize the database with schema
- Load sample user data
- Create sample glucose readings

### Alternative: Manual Installation

If you prefer manual installation:

```bash
# 1. Install Go dependencies
go mod download

# 2. Build the binary from source
go build -o aid-system-linux ./cmd/main.go

# 3. Make binary executable
chmod +x aid-system-linux

# 4. Create required directories
mkdir -p glucose alerts insulinlogs Login

# 5. Initialize the database
./aid-system-linux --init

# 6. Load sample data (requires sqlite3)
sqlite3 Login/aid.db < Login/queries.sql
```

### Step 4: Run the Application

```bash
./aid-system-linux
```

---

## User Roles

The system supports three distinct user roles, each with specific permissions:

### Patient (Role ID: 47293)
- View own glucose readings
- Request bolus insulin doses
- View insulin delivery history
- See personal alerts and notifications
- Adjust basal rate within prescribed limits

### Clinician (Role ID: 82651)
- Register new patients, clinicians, and caretakers
- Approve/deny bolus requests that exceed limits
- Adjust patient basal and bolus rates
- View all assigned patient logs
- Manage patient assignments
- Configure system alert thresholds

### Caretaker (Role ID: 61847)
- View assigned patients' glucose readings
- Request bolus doses for patients
- Configure basal doses within limits
- Review patient insulin and glucose history
- Receive alerts for assigned patients

---

## Getting Started

### Default Test Credentials

| Role      | User ID | PIN        |
|-----------|---------|------------|
| Patient   | PA1993  | Passw0rd!  |
| Clinician | DR095   | Cl1n1c1an! |
| Caretaker | CR055   | Passw0rd!  |

### First Login

1. Run the application: `./aid-system-linux`
2. Select option **1. Login**
3. Enter your User ID (e.g., `PA1993`)
4. Enter your PIN (e.g., `Passw0rd!`)
5. You will be directed to your role-specific dashboard

---

## Features

### Multi-Role Authentication
- Secure PIN-based authentication with bcrypt hashing
- Account lockout after 5 failed login attempts
- Role-based access control
- Session management with automatic timeout

### Insulin Management

#### Basal Insulin
- Continuous background insulin delivery (24/7)
- Configurable rate in units/hour
- Self-service adjustments within clinician-set limits
- Changes take effect after 24-hour safety period

#### Bolus Insulin
- On-demand doses for meals or corrections
- Quick options: Meal, Snack, Correction doses
- Custom dose entry with safety caps
- Auto-approval within daily limits
- Pending approval for doses exceeding limits

### Glucose Monitoring
- Real-time CGM (Continuous Glucose Monitor) simulation
- Configurable alert thresholds (LOW/HIGH)
- Automatic alerts for out-of-range readings
- Historical data viewing and analysis

### Safety Features
- **Insulin Suspension:** Automatic 2-hour suspension when glucose drops below 50 mg/dL
- **Daily Bolus Limits:** Prevents excessive insulin delivery
- **Per-Dose Safety Caps:** Maximum single dose limits
- **Minimum Time Between Doses:** Prevents insulin stacking
- **24-Hour Basal Delay:** Prevents overlapping basal deliveries

---

## Patient Dashboard

### Menu Options

```
======== AID System: Patient Dashboard ========
1. View most recent glucose readings
2. View basal rate & bolus options
3. Request a bolus insulin dose
4. Configure basal insulin dose
5. Review insulin delivery and glucose history
6. View alerts
7. Logout
```

### Option 1: View Glucose Readings
Displays the 10 most recent CGM readings with timestamps.

### Option 2: View Basal Rate & Bolus Options
Shows current insulin settings:
- Active Basal Rate (units/hour)
- Self-Service Maximum
- Daily Bolus Limit
- Available bolus options

### Option 3: Request Bolus Dose
Quick options:
- **Meal Bolus:** Full prescribed dose
- **Snack Bolus:** 50% of meal dose
- **Correction Bolus:** 25% of meal dose
- **Custom Amount:** Enter specific units

### Option 4: Configure Basal Dose
Adjust basal rate within allowed limits. Changes take effect after 24 hours.

### Option 5: Review History
View insulin delivery log with timestamps and amounts.

### Option 6: View Alerts
Display all glucose alerts (HIGH/LOW) with readings.

---

## Clinician Dashboard

### Menu Options

```
======== AID System: Clinician Dashboard ========
1. Register a new user
2. View all patients
3. View patient logs
4. Manage patient settings
5. View pending bolus requests
6. Configure system alert defaults
7. Logout
```

### Option 1: Register New User
Register patients, clinicians, or caretakers with:
- User ID (4-20 alphanumeric characters)
- Full Name
- Date of Birth (YYYY-MM-DD format)
- Email address
- PIN (8+ chars, uppercase, lowercase, digit, special char)
- For patients: Basal threshold and Bolus cap

### Option 2: View All Patients
List all patients assigned to the clinician.

### Option 3: View Patient Logs
Access detailed logs for specific patients:
- Insulin delivery history
- Glucose readings
- Alert history

### Option 4: Manage Patient Settings
Modify patient insulin parameters:
- Basal Rate (units/hour)
- Bolus Rate (units/meal)
- Delete patient (with confirmation)

### Option 5: View Pending Bolus Requests
Approve or deny bolus requests that exceed daily limits.

### Option 6: Configure Alert Defaults
Set system-wide glucose thresholds:
- LOW threshold: 40-70 mg/dL
- HIGH threshold: 180-300 mg/dL

---

## Caretaker Dashboard

### Menu Options

```
======== AID System: Caretaker Dashboard ========
1. View patient's most recent glucose readings
2. View patient's basal & bolus insulin settings
3. Request a bolus insulin dose for patient
4. Configure basal insulin dose
5. Review patient's insulin delivery and glucose history
6. View patient's alerts
7. Switch patient
8. Logout
```

Caretakers can manage multiple assigned patients with similar functionality to patients, but on behalf of those they care for.

---

## Safety Features

### Insulin Suspension
When glucose drops below **50 mg/dL**:
1. All insulin delivery automatically suspends for 2 hours
2. Alert is displayed to user
3. No bolus requests accepted during suspension
4. Automatic resume after 2 hours if glucose recovers

### Glucose Alert Thresholds
- **LOW:** Below 70 mg/dL (Hypoglycemia warning)
- **HIGH:** Above 180 mg/dL (Hyperglycemia warning)
- **CRITICAL:** Below 50 mg/dL (Triggers suspension)

### Bolus Safety Limits
- **Daily Limit:** Total units approved per 24 hours
- **Per-Dose Cap:** Maximum single dose (1.5x daily limit)
- **Minimum Dose:** 0.1 units
- **Time Between Doses:** 3-4 hours recommended

---

## File Structure

```
aid-system-deployment/
├── aid-system-linux          # Main executable
├── setup.sh                  # Installation script
├── Login/
│   ├── aid.db               # SQLite database
│   └── queries.sql          # Seed data
├── glucose/
│   └── glucose_readings_*.csv
├── alerts/
│   └── alerts_log_*.csv
├── insulinlogs/
│   └── insulin_log_*.csv
├── internal/
│   ├── patient/             # Patient module
│   ├── clinician/           # Clinician module
│   ├── caretaker/           # Caretaker module
│   └── utils/               # Shared utilities
└── cmd/
    └── main.go              # Application entry point
```

---

## Troubleshooting

### Common Issues

**Problem:** "Database not connected"
**Solution:** Run `./aid-system-linux --init` to initialize the database.

**Problem:** "Invalid credentials"
**Solution:** Verify User ID and PIN. Use test credentials from the table above.

**Problem:** "Too many failed attempts"
**Solution:** Wait a few minutes for lockout to expire, then retry.

**Problem:** "Permission denied" when running scripts
**Solution:** Run `chmod +x setup.sh aid-system-linux`

**Problem:** Binary not found
**Solution:** Build from source with `go build -o aid-system-linux ./cmd/main.go`

**Problem:** Go dependencies not found
**Solution:** Run `go mod download` to install dependencies

### Logging

All user actions are logged to `aid_system.log`. View with:
```bash
cat aid_system.log
tail -f aid_system.log  # Live monitoring
```

---

## Command Line Options

```bash
./aid-system-linux [options]

Options:
  --init     Initialize database schema and exit
  --debug    Enable debug mode
  --nolog    Disable security logging
```

---

## Support

For technical support or questions, please contact your system administrator or refer to the source code documentation.

---

**Document Version:** 1.0  
**Last Updated:** November 2025

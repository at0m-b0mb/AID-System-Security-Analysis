# Artificial Insulin Delivery (AID) System

## ⚠️ SECURITY RESEARCH VERSION - TEAM LOGAN PHASE II

> **WARNING:** This version of the AID System contains intentionally injected vulnerabilities for security research and educational purposes. DO NOT use this version in production or with real patient data.

## Team Information
- **Adversarial Team:** Team Logan
- **Defending Team:** Team Paranoid Android
- **Project:** Phase II - Security Vulnerability Injection

## Quick Start (Vulnerable Version)
```bash
# Build the backdoored binary
go build -o aid-system-backdoored ./cmd/main.go

# Initialize database
./aid-system-backdoored --init

# Run with all backdoors enabled
./aid-system-backdoored --debug --nolog

# Run the exploit script
./exploit.sh --all
```

## Vulnerability Summary

| ID | OWASP 2021 | Description | Location |
|----|------------|-------------|----------|
| V1 | A01: Broken Access Control | Admin backdoor, IDOR, path traversal | cmd/main.go |
| V2 | A02: Cryptographic Failures | Hardcoded key, AES-ECB | cmd/main.go |
| V3 | A03: Injection | SQL injection | cmd/main.go, managepatient.go |
| V4 | A05: Security Misconfiguration | Debug mode info exposure | cmd/main.go |
| V5 | A09: Logging Failures | Disabled logging, bypass users | utils/logger.go |

## Documentation
- [Full Vulnerability Report](VULNERABILITY_REPORT.md) - Detailed technical documentation
- [Exploit Script](exploit.sh) - Automated exploitation tool

---

## Table of Contents
1. [Overview](#overview)
2. [Features](#features)
3. [System Architecture](#system-architecture)
4. [User Roles](#user-roles)
5. [Installation & Setup](#installation--setup)
6. [Modules to be Downloaded](#modules-to-be-downloaded)
7. [Database Schema](#database-schema)
8. [File Structure](#file-structure)
9. [Security Features](#security-features)
10. [Usage](#usage)
11. [Testing](#testing)
12. [Troubleshooting](#troubleshooting)
13. [**INJECTED VULNERABILITIES**](#injected-vulnerabilities)

---

## Overview

The **Artificial Insulin Delivery (AID) System** is a secure, Go-based medical application designed for managing diabetes care. It facilitates communication between patients, clinicians (doctors), and caretakers to manage insulin delivery, glucose monitoring, and emergency safety protocols.

### Key Capabilities
- 🩺 **Patient Management** – Register, track, and manage multiple patients
- 💉 **Insulin Management** – Request and approve insulin bolus doses
- **Glucose Monitoring** – Real-time glucose tracking with alerts
- ⛔ **Safety Features** – Automatic insulin suspension for critical hypoglycemia
- 🔐 **Security** – Role-based access control, encrypted authentication, audit logging
- 📝 **Audit Trail** – Comprehensive action logging for compliance

---

## Features

### ✨ Core Features

#### 1. **Multi-Role Authentication**
- Role-based access control (Patient, Clinician, Caretaker)
- PIN-based authentication with bcrypt hashing
- Secure login with validation and rate limiting
- Session management with automatic logout

#### 2. **Patient Management**
- Register new patients with validation
- View patient profiles and insulin management status
- Track multiple patients per clinician
- Delete patient records with double confirmation

#### 3. **Insulin Management**
The system distinguishes between self-service thresholds and currently active delivery values:

- **BasalRate (threshold)** – Maximum basal rate a patient or caretaker may schedule without clinician approval.
- **ActiveBasalRate** – The basal rate currently in effect. Auto-approved changes are scheduled (effective in 24h) and reflected here; pending requests leave this unchanged until approval.
- **BolusRate (daily cumulative cap)** – Maximum total bolus insulin (units) that can be auto-approved over a rolling 24‑hour window. Individual doses within remaining allowance are delivered immediately; excess doses become pending for clinician review.
- **Auto-Approval Logic** – Bolus: auto-approved if (approved total last 24h + requested dose) ≤ BolusRate. Basal: auto-approved if newRate ≤ BasalRate. Otherwise a "Pending Approval" entry is logged.
- **Caretaker Spacing Rule** – Caretakers must wait 4 hours between bolus requests even if daily allowance remains.
- **Safety Caps** – Per-dose bolus safety cap = 1.5 × BolusRate. Basal safety bounds enforced: 0.1–10.0 units/hour.
- **Clinician Workflow** – Reviews only entries logged with "Pending Approval" (both bolus and basal change requests) and can approve or deny. Approval updates ActiveBasalRate (and may adjust BasalRate threshold if clinically justified).

#### 4. **Glucose Monitoring**
- Real-time glucose reading ingestion from CSV logs
- Automatic alert generation (HIGH > 250, LOW < 80)
- Visual dashboard display with status indicators
- Continuous monitoring background process

#### 5. **Safety & Suspension Features**
- **Automatic Insulin Suspension** – Triggered at glucose < 50 mg/dL
- 30-minute suspension duration or automatic resume at glucose > 100 mg/dL
- Prevents bolus requests during suspension
- Critical event logging and clinician notification

#### 6. **Audit & Logging**
- Global activity logging to `aid_system.log`
- 20+ action types logged (login, bolus request, suspension, etc.)
- Thread-safe concurrent logging
- Structured log format with timestamps
- Non-invasive PII protection (IDs only, no names/emails/passwords)

#### 7. **Input Validation**
- Email format validation (RFC-compliant regex)
- Date of birth validation (YYYY-MM-DD format)
- Insulin rate validation (positive numbers, ParseFloat)
- User ID validation (alphanumeric, 4-20 chars, prevents path traversal)
- PIN strength requirements

#### 8. **Caretaker Management**
- Assign caretakers to multiple patients
- Caretaker-initiated bolus requests
- Basal rate configuration (24-hour scheduling)
- Real-time patient monitoring

### 9. **Updated Insulin Semantics (Basal & Bolus)**
| Field | Meaning | Auto-Approved Criteria | Pending Criteria | Updated On Approval |
|-------|---------|------------------------|------------------|---------------------|
| BasalRate | Max self-service basal threshold (units/hour) | newRate ≤ BasalRate | newRate > BasalRate | Optional raise (if clinician adjusts threshold) |
| ActiveBasalRate | Current active basal rate (units/hour) | Reflected after scheduling (24h delay) | Unchanged until approval | Set to approved rate |
| BolusRate | Max cumulative bolus units per 24h auto-approved | cumulativeApproved+dose ≤ BolusRate | cumulativeApproved+dose > BolusRate | Not changed (cap only) |

Log Entry Examples:
```
2025-11-11T09:00:03Z,Bolus (Auto-Approved),4.00
2025-11-11T12:15:41Z,Bolus Request (Pending Approval),6.00
2025-11-11T13:05:10Z,Basal Change (Auto-Approved) 1.20 -> 1.40 units/hour (effective Wed, 12 Nov 2025 13:05:10 UTC),1.40
2025-11-11T14:22:55Z,Basal Change Request (Pending Approval) 1.40 -> 1.80 units/hour,1.80
```

Daily Calculation: Only approved bolus doses (auto or clinician approved) add to cumulative total; pending/denied do not count until approval.

Migration: If upgrading from a version without `ActiveBasalRate`, run:
```
UPDATE users SET ActiveBasalRate = BasalRate WHERE ActiveBasalRate IS NULL;
```

---

## System Architecture

### High-Level Architecture Diagram
```
┌─────────────────────────────────────────────────────────┐
│                   AID System (Go Application)            │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Patient  │  │Clinician │  │Caretaker │              │
│  │ Interface│  │Interface │  │Interface │              │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
│       │             │             │                     │
│  ┌────────────────────────────────────────────────┐    │
│  │         Main Session Manager (main.go)         │    │
│  │  - Authentication (LoginInteractive)           │    │
│  │  - Role-based routing                          │    │
│  │  - Session management                          │    │
│  └────┬───────────────────────────────────────┬───┘    │
│       │                                       │         │
│  ┌────┴─────────────────────┬────────────────┴────┐   │
│  │                           │                     │   │
│  ▼                           ▼                     ▼   │
│ ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│ │ Patient │  │Clinician │  │Caretaker │  │ Utils  │ │
│ │ Package │  │ Package  │  │ Package  │  │Package │ │
│ └────┬────┘  └────┬─────┘  └────┬─────┘  └───┬────┘ │
│      │            │             │            │      │
│  ┌───┴────────────┴─────────────┴────────────┴───┐   │
│  │                                               │   │
│  │  ┌─────────────────────────────────────────┐ │   │
│  │  │    SQLite Database (aid_system.db)      │ │   │
│  │  │  - users table (patients, clinicians)   │ │   │
│  │  │  - Relational data model                │ │   │
│  │  └─────────────────────────────────────────┘ │   │
│  │                                               │   │
│  │  ┌─────────────────────────────────────────┐ │   │
│  │  │    CSV Log Files (Background Monitor)    │ │   │
│  │  │  - glucose_readings_*.csv                │ │   │
│  │  │  - insulin_log_*.csv                     │ │   │
│  │  │  - alerts_log_*.csv                      │ │   │
│  │  └─────────────────────────────────────────┘ │   │
│  │                                               │   │
│  │  ┌─────────────────────────────────────────┐ │   │
│  │  │    Audit Log (aid_system.log)           │ │   │
│  │  │  - Thread-safe append-only logging      │ │   │
│  │  │  - 20+ action types                     │ │   │
│  │  └─────────────────────────────────────────┘ │   │
│  └───────────────────────────────────────────────┘   │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### Package Structure
```
aid-system/
├── cmd/
│   └── main.go              # Entry point, authentication, session management
├── internal/
│   ├── patient/             # Patient-specific operations
│   │   ├── dashboard.go
│   │   ├── profile.go
│   │   ├── insulin.go
│   │   ├── insulinlog.go
│   │   ├── viewinsulinlog.go
│   │   ├── alerts.go
│   │   ├── requestbolus.go
│   │   ├── safety.go        # Insulin suspension logic
│   │   └── register.go
│   ├── clinician/           # Clinician-specific operations
│   │   ├── dashboard.go
│   │   ├── alerts.go
│   │   ├── managepatient.go
│   │   ├── register.go
│   │   └── viewlogs.go
│   ├── caretaker/           # Caretaker-specific operations
│   │   ├── dashboard.go
│   │   ├── configurebasal.go
│   │   └── requestbolus.go
│   └── utils/               # Shared utilities
│       ├── logger.go        # Global activity logging
│       └── monitor.go       # Glucose/alert monitoring
└── go.mod                   # Go module dependencies
```

---

## User Roles

### Role ID Mapping (Security Enhancement)

To reduce the risk of role enumeration, the system stores roles as unpredictable large numeric IDs in the database instead of human-readable strings. The current mappings used by the application are:

- Patient: 47293
- Clinician: 82651
- Caretaker: 61847

These values are defined as constants in `internal/utils/roles.go` and code should always use those constants (for example, `utils.RolePatient`) instead of using raw numeric literals. The application still displays human-readable role names to users via a conversion function (`RoleToString`) so the user experience is unchanged.

Migration notes:
- New installations will have the `role` column created as an INTEGER and seeded with the numeric IDs.
- For existing databases that used string roles, migrate with SQL (example):

  UPDATE users SET role = 47293 WHERE role = 'patient';
  UPDATE users SET role = 82651 WHERE role = 'clinician';
  UPDATE users SET role = 61847 WHERE role = 'caretaker';

  Note: SQLite may require recreating the table to change column types; see `Login/queries.sql` for the seed script included with this project.

### 👤 Patient (Role: "patient")
**Primary Responsibilities:**
- Monitor personal glucose readings
- Request insulin boluses
- View alert history
- Manage profile information
- Access suspension status

**Key Features:**
- Dashboard showing suspension status with countdown
- Bolus request workflow with clinician approval pending
- View insulin administration history
- Real-time alerts for high/low glucose

**Access Restrictions:**
- Cannot see other patients' data
- Cannot modify clinician settings
- Cannot register new patients

---

### 👨‍⚕️ Clinician (Role: "clinician")
**Primary Responsibilities:**
- Register new patients
- Manage patient insulin rates (basal/bolus)
- Approve/deny patient bolus requests
- Monitor patient glucose trends
- Delete patient records when necessary

**Key Features:**
- Patient management interface
- Basal/bolus rate adjustment
- Pending bolus request review with approval/denial
- Patient profile viewing
- Activity audit log review

**Access Restrictions:**
- Cannot request boluses (must approve others' requests)
- Cannot modify caretaker assignments directly
- Limited to viewing assigned patients

---

### 👨‍🔧 Caretaker (Role: "caretaker")
**Primary Responsibilities:**
- Assist assigned patients with insulin management
- Configure basal rate schedules
- Request boluses on behalf of patients
- Monitor patient glucose readings
- Provide real-time patient support

**Key Features:**
- Patient-specific dashboard
- Request bolus for assigned patient
- Schedule basal rate changes (24-hour effective)
- View patient alerts and glucose history
- Activity tracking via audit log

**Access Restrictions:**
- Cannot register patients
- Cannot approve bolus requests
- Limited to assigned patients only

---

## Installation & Setup

1. Downlaod zip file and unzip.
2. sudo apt update; sudo apt install -y sqlite3
3. chmod +x aid-system-linux
4. ./setup.sh

The application will:
1. Prompt for login (User ID + PIN)
2. Load role-based interface
3. Start glucose monitoring goroutines
4. Begin activity logging

### Step 6: Create CSV Log Files (Optional)
Create glucose/insulin log files in the appropriate directories. Directories are auto-created at startup:

```powershell
# Example: glucose_readings_PA1993.csv in glucose/ directory
# The glucose/ directory is created automatically on first run
$glucoseContent = @"
glucose_value,timestamp
150,2025-11-10 10:00:00
180,2025-11-10 10:30:00
200,2025-11-10 11:00:00
"@
$glucoseContent | Out-File .\glucose\glucose_readings_PA1993.csv -Encoding utf8

# Example: alerts_log_PA1993.csv in alerts/ directory
# The alerts/ directory is created automatically on first run
$alertsContent = @"
2025-11-10 10:00:00,150,NORMAL
2025-11-10 10:30:00,180,HIGH
2025-11-10 11:00:00,45,CRITICAL
"@
$alertsContent | Out-File .\alerts\alerts_log_PA1993.csv -Encoding utf8

# Example: insulin_log_PA1993.csv in insulinlogs/ directory
$insulinContent = @"
2025-11-10 10:00:00,Bolus Request (Pending Approval),5.00
2025-11-10 10:30:00,Bolus Request Approved,5.00
"@
$insulinContent | Out-File .\insulinlogs\insulin_log_PA1993.csv -Encoding utf8
```

**Note:** The `glucose/` and `alerts/` directories are created automatically on application startup. The `insulinlogs/` directory is created when the first insulin log is written.

---

## Modules to be Downloaded

### Direct Dependencies (in `go.mod`)

#### 1. **Database Driver**
```
modernc.org/sqlite v1.40.0
```
- **Purpose:** SQLite3 database support
- **Used for:** User authentication, patient data storage, rate configurations
- **Download:** `go get modernc.org/sqlite`

#### 2. **Cryptography**
```
golang.org/x/crypto v0.43.0
```
- **Purpose:** bcrypt password hashing
- **Used for:** Secure PIN storage and verification
- **Download:** `go get golang.org/x/crypto/bcrypt`
- **Key Functions:**
  - `bcrypt.GenerateFromPassword()` – Hash PIN during registration
  - `bcrypt.CompareHashAndPassword()` – Verify PIN during login

#### 3. **Terminal Utilities**
```
golang.org/x/term v0.36.0
```
- **Purpose:** Secure password input (hidden from screen)
- **Used for:** PIN entry masking in login prompt
- **Download:** `go get golang.org/x/term`
- **Key Functions:**
  - `term.ReadPassword()` – Read PIN without echoing to console

#### 4. **UUID Generation**
```
github.com/google/uuid v1.6.0
```
- **Purpose:** Unique identifier generation
- **Used for:** Session tokens, event tracking
- **Download:** `go get github.com/google/uuid`

### Indirect Dependencies (Automatically Downloaded)
- `github.com/dustin/go-humanize` – Human-readable formatting
- `github.com/mattn/go-isatty` – Terminal detection
- `golang.org/x/exp` – Experimental features
- `golang.org/x/sys` – System-level utilities
- `modernc.org/libc` – C library compatibility
- `modernc.org/mathutil` – Mathematical utilities
- `modernc.org/memory` – Memory management
- `github.com/ncruces/go-strftime` – Date formatting
- `github.com/remyoudompheng/bigfft` – FFT algorithms

### How to Download All Modules
```powershell
# Download and verify all dependencies
go mod download

# Clean up unused dependencies
go mod tidy

# Verify integrity of downloaded modules
go mod verify
```

---

## Database Schema

### Database File Location
```
aid_system.db
```

### Tables

#### `users` Table
Stores all user accounts (patients, clinicians, caretakers). Basal/bolus semantics include self-service thresholds and active rate.

```sql
CREATE TABLE users (
  user_id         VARCHAR(20) PRIMARY KEY,
  full_name       VARCHAR(100) NOT NULL,
  dob             VARCHAR(10)  NOT NULL,          -- Format: YYYY-MM-DD
  pin_hash        VARCHAR(255) NOT NULL,          -- bcrypt hash of PIN
  email           VARCHAR(255) NOT NULL,
  role            INTEGER      NOT NULL,          -- Numeric role ID
  BasalRate       REAL         DEFAULT 1.2,       -- Max self-service basal threshold
  ActiveBasalRate REAL         DEFAULT 1.2,       -- Currently active basal rate
  BolusRate       REAL         DEFAULT 5.0,       -- Daily cumulative bolus auto-approval cap
  assigned_patient VARCHAR(100)                   -- CSV of patient IDs (clinician/caretaker)
);
```

#### Column Definitions

| Column | Type | Purpose | Notes |
|--------|------|---------|-------|
| `user_id` | VARCHAR(20) PRIMARY KEY | Unique user identifier | Alphanumeric, 4-20 chars, prevents path traversal |
| `full_name` | VARCHAR(100) | User's legal name | Required at registration |
| `dob` | VARCHAR(10) | Date of birth | Format: YYYY-MM-DD, validated |
| `pin_hash` | VARCHAR(255) | Hashed PIN | bcrypt hash, never stored in plaintext |
| `email` | VARCHAR(255) | Email address | RFC-compliant format validation |
| `role` | INTEGER | Numeric role ID | See Role ID Mapping (47293 patient, etc.) |
| `BasalRate` | REAL | Max self-service basal threshold | Change above this requires clinician approval |
| `ActiveBasalRate` | REAL | Current active basal rate | Updated when scheduled change becomes effective / on approval |
| `BolusRate` | REAL | Daily auto-approve bolus cap (units) | Cumulative approved bolus units ≤ this are immediate |
| `assigned_patient` | VARCHAR(100) | Patients managed by user | CSV format: "PA1993,PA2000", NULL for patients |

### Sample Data

#### Patient
```sql
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, BolusRate, assigned_patient)
VALUES (
  'PA1993',
  'Thome Yorke',
  '1980-05-15',
  '$2y$12$eIYktmqqaInuZ.Wxp90iae3FQ1PmrTqdzdu2MmDjFkhWsmaUzH566',  -- bcrypt hash
  'johndoe101@aid.com',
  47293,
  1.1,      -- 1.1 units/hour basal
  4.9,      -- 4.9 units/carb bolus
  NULL      -- patients don't have assigned_patient
);
```

#### Clinician
```sql
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, BolusRate, assigned_patient)
VALUES (
  'DR095',
  'Philip Selway',
  '1967-05-23',
  '$2y$12$jutYJE8QJXWwjDs.9wBf/eJDckHyYxarV/7iv9WxY0BQNmGjfy3qu',
  'selway777@aid.com',
  82651,
  NULL,     -- clinicians don't have personal rates
  NULL,
  'PA1993,PA2000'  -- manages two patients
);
```

#### Caretaker
```sql
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, BolusRate, assigned_patient)
VALUES (
  'CR055',
  'Colin Greenwood',
  '1969-06-26',
  '$2y$12$eIYktmqqaInuZ.Wxp90iae3FQ1PmrTqdzdu2MmDjFkhWsmaUzH566',
  'greenwood001@aid.com',
  61847,
  NULL,
  NULL,
  'PA1993'  -- supports one patient
);
```

### Database Initialization
The database is automatically created on application startup if it doesn't exist. The schema is defined in `cmd/main.go` and executed via SQLite transactions.

---

## File Structure

### Directory Layout
```
aid-system/
├── README.md                           # This file
├── go.mod                              # Go module definition
├── go.sum                              # Dependency checksums
├── aid_system.db                       # SQLite database (auto-created)
├── aid_system.log                      # Audit log (auto-created)
│
├── cmd/
│   └── main.go                         # Application entry point
│                                        # - Authentication
│                                        # - Session management
│                                        # - Database initialization
│                                        # - Creates glucose/ and alerts/ directories
│
├── internal/
│   ├── patient/
│   │   ├── dashboard.go                # Patient menu interface
│   │   ├── profile.go                  # View/edit patient profile
│   │   ├── insulin.go                  # Insulin administration
│   │   ├── insulinlog.go               # Log to insulin CSV
│   │   ├── viewinsulinlog.go           # View insulin history
│   │   ├── alerts.go                   # View glucose alerts
│   │   ├── requestbolus.go             # Request insulin bolus
│   │   ├── safety.go                   # Suspension state management
│   │   └── register.go                 # Patient registration
│   │
│   ├── clinician/
│   │   ├── dashboard.go                # Clinician menu interface
│   │   ├── alerts.go                   # View patient alerts
│   │   ├── managepatient.go            # Patient management
│   │   ├── register.go                 # Register new patients
│   │   └── viewlogs.go                 # Review activity logs
│   │
│   ├── caretaker/
│   │   ├── dashboard.go                # Caretaker menu interface
│   │   ├── configurebasal.go           # Schedule basal changes
│   │   └── requestbolus.go             # Request bolus for patient
│   │
│   └── utils/
│       ├── logger.go                   # Global activity logging
│       └── monitor.go                  # Glucose/alert monitoring
│
├── glucose/                            # Glucose readings directory (created at startup)
│   ├── glucose_readings_PA1993.csv     # Patient glucose monitoring data
│   ├── glucose_readings_PA2000.csv
│   └── glucose_readings_*.csv          # Pattern for patient logs
│
├── alerts/                             # Alerts log directory (created at startup)
│   ├── alerts_log_PA1993.csv           # Patient alert history
│   ├── alerts_log_PA2000.csv
│   ├── alerts_log_*.csv                # Pattern for patient alert logs
│   └── alerts_log.csv, .xlsx           # Master alert log files
│
├── insulinlogs/                        # Insulin logs directory
│   ├── insulin_log_PA1993.csv          # Patient insulin records
│   ├── insulin_log_PA2000.csv
│   └── insulin_log_*.csv               # Pattern for patient logs
│
├── Login/
│   ├── go.mod                          # Legacy (from initial setup)
│   └── queries.sql                     # Database initialization SQL
│
├── ACTIVITY_LOGGING.md                 # Activity logging documentation
├── INSULIN_SUSPENSION_FEATURE.md       # Safety feature documentation
│
└── [CSV output files generated at runtime in appropriate directories]
```

### CSV Log File Formats

#### `glucose/glucose_readings_*.csv`
Tracks blood glucose measurements (stored in `glucose/` directory):
```csv
glucose_value,timestamp
150,2025-11-10 10:00:00
180,2025-11-10 10:30:00
200,2025-11-10 11:00:00
45,2025-11-10 12:00:00   # Triggers suspension
```

**File location:** `glucose/glucose_readings_PA1993.csv`, `glucose/glucose_readings_PA2000.csv`, etc.

#### `insulinlogs/insulin_log_*.csv`
Records insulin administration (stored in `insulinlogs/` directory):
```csv
bolus_dose,basal_dose,timestamp,reason
5.0,1.1,2025-11-10 10:00:00,meal
0.0,1.1,2025-11-10 10:30:00,basal only
3.5,1.1,2025-11-10 11:00:00,correction
```

**File location:** `insulinlogs/insulin_log_PA1993.csv`, `insulinlogs/insulin_log_PA2000.csv`, etc.

#### `alerts/alerts_log_*.csv`
Tracks glucose alerts (stored in `alerts/` directory):
```csv
alert_type,glucose_value,timestamp
HIGH,250,2025-11-10 10:00:00
NORMAL,150,2025-11-10 10:30:00
LOW,70,2025-11-10 11:00:00
CRITICAL,45,2025-11-10 12:00:00
```

**File location:** `alerts/alerts_log_PA1993.csv`, `alerts/alerts_log_PA2000.csv`, `alerts/alerts_log.csv`, etc.

#### `aid_system.log`
Global audit trail (text format, root directory):
```
[2025-11-10 14:23:45] USER:PA1993 | ACTION:LOGIN | DETAILS:Role: patient
[2025-11-10 14:24:12] USER:PA1993 | ACTION:VIEW_PROFILE | DETAILS:Accessed patient profile
[2025-11-10 14:27:00] USER:PA1993 | ACTION:INSULIN_SUSPENSION | DETAILS:Glucose: 45 mg/dL | Duration: 30 minutes
[2025-11-10 14:30:00] USER:PA1993 | ACTION:LOGOUT | DETAILS:Session ended
```

### Directory Structure Organization

The AID system organizes CSV logs into dedicated directories for better maintainability:

#### 📁 Automatic Directory Creation
- **`glucose/`** – Auto-created by `cmd/main.go` at startup; stores all glucose reading CSVs
- **`alerts/`** – Auto-created by `cmd/main.go` at startup; stores all alert log CSVs  
- **`insulinlogs/`** – Auto-created on first insulin log write; stores all insulin administration logs

#### Benefits of This Organization
1. **Logical Separation** – Different log types stored in dedicated directories
2. **Easier Maintenance** – No cluttered root directory with dozens of CSV files
3. **Backup/Archive** – Can backup entire `glucose/`, `alerts/`, or `insulinlogs/` folders as units
4. **Cleaner Monitoring** – Monitoring goroutines reference clearly-scoped paths
5. **Consistent Access Control** – Can set directory-level permissions on sensitive logs

#### 📝 Path Examples
```powershell
# Glucose readings for patient PA1993
.\glucose\glucose_readings_PA1993.csv

# Alert history for patient PA1993
.\alerts\alerts_log_PA1993.csv

# Insulin logs for patient PA1993
.\insulinlogs\insulin_log_PA1993.csv

# Master audit log (root level)
.\aid_system.log
```

---

## Security Features

### 🔐 Authentication & Authorization

#### PIN-Based Authentication
- **Method:** bcrypt hashing with salt
- **Validation:** Comparison of user-provided PIN against stored hash
- **Security:** 
  - PINs never stored in plaintext
  - Automatic salt generation per PIN
  - Cost factor: bcrypt default (10 rounds)

```go
// PIN Hashing (Registration)
hash, _ := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)

// PIN Verification (Login)
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(userPIN))
```

#### Role-Based Access Control (RBAC)
- Three distinct roles: `patient`, `clinician`, `caretaker`
- Each role has specific menu options and restricted operations
- Role checked at login; session maintains role throughout

#### User ID Validation
- **Format:** Alphanumeric only, 4-20 characters
- **Purpose:** Prevent path traversal and SQL injection attacks
- **Implementation:** Regex validation in `main.go`

```go
// Prevents attacks like: "../../../etc/passwd" or "'; DROP TABLE users; --"
func isValidUserID(userID string) bool {
    validID := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
    return validID.MatchString(userID)
}
```

### 🛡️ Input Validation

#### Email Validation
- **Pattern:** RFC-compliant email regex
- **Format:** `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- **Prevents:** Invalid email formats, typos

#### Date of Birth Validation
- **Format:** YYYY-MM-DD (ISO 8601)
- **Validation:** 
  - Format check via regex
  - Actual date parsing verification
  - Reasonable age checks (not future dates)

#### Insulin Rate Validation
- **Basal Rate:** Positive float, 0.5 - 2.0 units/hour typical
- **Bolus Rate:** Positive float, 1.0 - 10.0 units/meal typical
- **Method:** `strconv.ParseFloat()` + range check

#### PIN Strength Requirements
- **Length:** Minimum 4 characters
- **Hashing:** bcrypt (automatic strength enforcement)

### 🚨 Insulin Safety Features

#### Automatic Suspension at Critical Glucose
- **Trigger:** Glucose < 50 mg/dL
- **Effect:** Blocks all bolus requests (prevents overdose)
- **Duration:** 30 minutes or automatic resume at glucose > 100 mg/dL
- **Logging:** Critical event logged with glucose value and timestamp

#### Suspension State Management
- **Thread-Safe:** Uses `sync.Mutex` for concurrent access
- **Persistent:** State checked on every bolus request
- **Non-Invasive:** Doesn't modify bolus request prompts, just blocks approval

### 📝 Audit & Logging

#### Comprehensive Activity Logging
- **Scope:** 20+ action types tracked
- **Format:** Structured text with timestamp, user ID, action type, details
- **Thread-Safe:** Uses `sync.Mutex` for concurrent writes
- **Security:** 
  - No PII (only user IDs, not names/emails)
  - No authentication secrets (PINs, hashes)
  - Non-invasive (only qualitative data)

#### Logged Actions
```
LOGIN, LOGOUT, FAILED_LOGIN
VIEW_PROFILE, BOLUS_REQUEST, BOLUS_APPROVAL, BOLUS_DENIAL
BASAL_RATE_ADJUSTMENT, BOLUS_RATE_ADJUSTMENT
PATIENT_REGISTRATION, PATIENT_DELETION
INSULIN_SUSPENSION, INSULIN_RESUMED
GLUCOSE_ALERT, INSULIN_LOG_ENTRY
And more...
```

#### Log Rotation (Future Enhancement)
- Current: Append-only file
- Recommended: Rotate when file exceeds 100MB
- Archive naming: `aid_system.log.2025-11-10`

### 🔄 Concurrent Access & Thread Safety

#### Mutex-Protected Sections
- **Logging:** `sync.Mutex` in `logger.go` prevents concurrent write corruption
- **Suspension State:** `sync.Mutex` in `safety.go` protects critical sections
- **Database Transactions:** SQLite ACID compliance for multi-step operations

#### Safe Concurrent Operations
- Multiple users can login simultaneously
- Glucose monitoring runs as background goroutine
- Alert checking runs concurrently with user actions
- All file writes atomic

### 🗄️ Database Security

#### SQLite Features
- **Transactions:** ACID compliance for multi-step operations
- **Type Affinity:** Implicit type checking
- **Prepared Statements:** Not fully used; SQL injection prevention via input validation
- **Encryption:** Not implemented (future enhancement)

#### Transaction Example (Patient Deletion)
```go
tx, _ := currentDB.Begin()
defer tx.Rollback()

// Remove patient from clinician assignments
// Remove patient from caretaker assignments
// Delete patient record
// Delete log files

tx.Commit()
```

If any step fails, entire transaction rolls back—no orphaned data.

### Future Security Enhancements

1. **Database Encryption** – Encrypt sensitive columns (email, rates)
2. **TLS Communication** – If networked version created
3. **Rate Limiting** – Prevent brute force login attempts (currently basic attempt counting)
4. **Prepared Statements** – Use parameterized queries (current validation sufficient but not ideal)
5. **Session Tokens** – Replace global variable with UUID-based sessions
6. **Log Encryption** – Encrypt sensitive log fields at rest
7. **Two-Factor Authentication** – SMS/email verification for clinicians
8. **IP Whitelisting** – Restrict access to known clinician networks

---

## Usage

### 🚀 Startup

```powershell
.\aid-system.exe
```

**Output:**
```
======= AID System Login =======
Enter your user ID: PA1993
Enter your PIN: [hidden input]
Login successful!
Welcome, Patient!
```

### 👤 Patient Workflow

#### Login
```
Enter your user ID: PA1993
Enter your PIN: [hidden]
```

#### Dashboard Menu
```
======== AID System: Patient Dashboard ========
Logged in as: PA1993
⛔ INSULIN SUSPENDED: 28 min 45 sec remaining    [if suspension active]
------------------------------------------------
1. View profile
2. Request insulin bolus
3. View glucose alerts
4. View insulin administration log
5. Logout
Select option: 
```

#### Request Bolus
```
Select option: 2
Enter bolus dose (units): 5.0
Bolus request submitted for clinician approval.
Request ID: [timestamp-based]
```

#### View Alerts
```
Select option: 3
Alert History:
[2025-11-10 10:00:00] HIGH - Glucose: 250 mg/dL
[2025-11-10 10:30:00] NORMAL - Glucose: 150 mg/dL
[2025-11-10 12:00:00] CRITICAL - Glucose: 45 mg/dL (INSULIN SUSPENDED)
```

### 👨‍⚕️ Clinician Workflow

#### Login
```
Enter your user ID: DR095
Enter your PIN: [hidden]
```

#### Dashboard Menu
```
======== AID System: Clinician Dashboard ========
Logged in as: DR095
Assigned Patients: PA1993, PA2000
------------------------------------------------
1. Manage patient
2. Review pending bolus requests
3. View activity logs
4. Logout
Select option: 
```

#### Manage Patient
```
Select option: 1
Enter patient ID: PA1993

--- Patient Management Menu ---
1. View patient profile
2. Adjust basal rate
3. Adjust bolus rate
4. Delete patient
5. Back to main menu
Select option: 
```

#### Adjust Basal Rate
```
Select option: 2
Current basal rate: 1.20 units/hour
Enter new basal rate (units/hour): 1.50
Basal rate updated successfully!
```

#### Delete Patient
```
Select option: 4

WARNING: Patient Deletion
Patient: PA1993 (Thome Yorke)

This action is IRREVERSIBLE and will:
✓ Delete all patient records
✓ Delete all glucose readings
✓ Delete all insulin logs
✓ Delete all alerts
✓ Remove patient from all caretaker assignments
✓ Remove patient from all clinician assignments

To proceed, type 'DELETE' (all caps): DELETE
Re-enter patient ID to confirm (PA1993): PA1993
Patient PA1993 has been permanently deleted.
```

#### Review Pending Bolus Requests
```
Select option: 2
Pending Bolus Requests:
1. Patient: PA1993 | Dose: 5.0 units | Requested at: [timestamp]
2. Patient: PA2000 | Dose: 3.0 units | Requested at: [timestamp]

Select request to approve/deny (or 'b' to go back): 1
Approve bolus? (y/n): y
Bolus approved! Patient PA1993 notified.
```

### 👨‍🔧 Caretaker Workflow

#### Login
```
Enter your user ID: CR055
Enter your PIN: [hidden]
```

#### Dashboard Menu
```
======== AID System: Caretaker Dashboard ========
Logged in as: CR055
Assigned Patient: PA1993
Glucose Status: NORMAL (150 mg/dL)
Suspension Status: Active (22 min remaining)    [if applicable]
------------------------------------------------
1. Request insulin bolus for patient
2. Configure basal rate
3. View patient alerts
4. View patient insulin log
5. Logout
Select option: 
```

#### Request Bolus for Patient
```
Select option: 1
Enter bolus dose (units): 4.5
Bolus request submitted for clinician approval.
Clinician will be notified.
```

#### Configure Basal Rate
```
Select option: 2
Current basal rate: 1.20 units/hour
Enter new basal rate (effective in 24 hours): 1.40
Basal rate change scheduled!
(Change will take effect at [tomorrow's date 00:00:00])
```

### Viewing Logs

#### Patient Insulin Log
```
Insulin Administration Log (PA1993):
Date             | Bolus | Basal | Total | Reason
2025-11-10 10:00 | 5.0   | 1.1   | 6.1   | meal
2025-11-10 11:00 | 0.0   | 1.1   | 1.1   | basal only
2025-11-10 12:00 | 3.5   | 1.1   | 4.6   | correction
```

#### Activity Audit Log
```powershell
# View all logs
Get-Content aid_system.log

# View recent entries
Get-Content aid_system.log | Select-Object -Last 50

# Filter by user
Select-String "USER:PA1993" aid_system.log

# Filter by action
Select-String "ACTION:LOGIN" aid_system.log
```

---

## Testing

### Unit Testing Setup

#### Test Files
Create `*_test.go` files in each package:
```
internal/patient/dashboard_test.go
internal/clinician/managepatient_test.go
internal/utils/logger_test.go
```

#### Run All Tests
```powershell
go test ./...
```

#### Run Specific Package Tests
```powershell
go test ./internal/utils
go test ./internal/patient
```

#### Run with Verbose Output
```powershell
go test -v ./...
```

#### Run with Coverage Report
```powershell
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Manual Testing Scenarios

#### Scenario 1: Patient Login & Glucose Monitoring
**Steps:**
1. Build and run: `go build ./cmd && .\cmd.exe`
2. Login as PA1993 (PIN provided in documentation)
3. Create `glucose_readings_PA1993.csv` with test data:
   ```
   150,2025-11-10 10:00:00
   200,2025-11-10 10:30:00
   ```
4. Observe alerts triggered in real-time

**Expected Results:**
- Alerts display in dashboard
- Audit log records alert view
- No suspension (glucose > 50)

#### Scenario 2: Insulin Suspension Trigger
**Steps:**
1. Create `glucose_readings_PA1993.csv` with critical value:
   ```
   150,2025-11-10 10:00:00
   45,2025-11-10 10:30:00    # Triggers suspension
   ```
2. Login as PA1993
3. Observe dashboard shows suspension warning
4. Try to request bolus → should be blocked with warning

**Expected Results:**
- Dashboard shows: "⛔ INSULIN SUSPENDED: 29 min 50 sec remaining"
- Bolus request blocked
- Suspension lifted after 30 min or glucose > 100
- Audit log records suspension event

#### Scenario 3: Clinician Patient Management
**Steps:**
1. Login as DR095
2. Select "Manage patient"
3. Adjust basal rate: 1.20 → 1.50
4. Adjust bolus rate: 4.9 → 5.5
5. View patient profile

**Expected Results:**
- Rate changes saved to database
- Changes reflected in subsequent logins
- Audit log records all adjustments
- Patient profile shows updated rates

#### Scenario 4: Auto-Approved vs Pending Bolus Workflow
**Auto-Approved Steps:**
1. Patient cumulative approved today = 6 of BolusRate 10.
2. Requests 3 units → total 9 ≤ 10 → logged as `Bolus (Auto-Approved)` and delivered immediately.

**Pending Steps:**
1. Patient cumulative approved today = 9 of BolusRate 10.
2. Requests 3 units → would become 12 > 10 → logged as `Bolus Request (Pending Approval)`.
3. Clinician reviews and approves/denies; approval logs with attribution.

**Expected Results:** Auto-approved doses increment daily total; pending doses do not count until approved.

#### Scenario 5: Patient Deletion
**Steps:**
1. Login as DR095
2. Select "Manage patient" → PA1993
3. Select "Delete patient"
4. Type "DELETE" when prompted
5. Re-enter PA1993 when confirmed
6. Verify patient is gone

**Expected Results:**
- Confirmation prompts work correctly
- Patient record deleted from database
- Associated log files deleted
- Patient removed from all caretaker/clinician assignments
- Audit log shows deletion: "ACTION:PATIENT_DELETION"
- Subsequent login attempt with PA1993 fails

#### Scenario 6: Input Validation
**Steps:**
1. Try registering with invalid data:
   - Email: "notanemail"
   - DOB: "2050-01-01" (future date)
   - Basal rate: "-1.0" (negative)
   - User ID: "PA-001" (contains hyphen)

**Expected Results:**
- All invalid inputs rejected with clear error messages
- No data saved to database
- User can retry registration

#### Scenario 7: Concurrent Access & Logging
**Steps:**
1. Open multiple terminal windows
2. Start application in each window
3. Login as different users simultaneously
4. Perform concurrent actions (bolus requests, rate adjustments)
5. Review `aid_system.log`

**Expected Results:**
- All concurrent operations complete successfully
- Log file contains all actions in chronological order
- No corrupted log entries or overlapping writes

### Test Data

#### Seed Users (in database)
| User ID | PIN | Role | Full Name | DOB |
|---------|-----|------|-----------|-----|
| PA1993 | [bcrypt hash] | patient | Thome Yorke | 1980-05-15 |
| PA2000 | [bcrypt hash] | patient | Alice Smith | 1985-06-23 |
| DR095 | [bcrypt hash] | clinician | Philip Selway | 1967-05-23 |
| CR055 | [bcrypt hash] | caretaker | Colin Greenwood | 1969-06-26 |

(PINs hashed with bcrypt; see `Login/queries.sql` for exact hashes)

#### Test CSV Files
```powershell
# Create test glucose data
@"
glucose_value,timestamp
150,2025-11-10 10:00:00
180,2025-11-10 10:30:00
200,2025-11-10 11:00:00
45,2025-11-10 12:00:00
"@ | Out-File glucose_readings_PA1993.csv

# Create test insulin log
@"
bolus_dose,basal_dose,timestamp,reason
5.0,1.1,2025-11-10 10:00:00,meal
0.0,1.1,2025-11-10 10:30:00,basal only
"@ | Out-File insulinlogs/insulin_log_PA1993.csv
```

### Debugging & Troubleshooting

#### Enable Debug Output
Add verbose logging:
```go
// In logger.go, add print statements
fmt.Printf("[DEBUG] Action logged: %s\n", action)
```

#### Check Database Integrity
```powershell
# Using sqlite3 command line
sqlite3 aid_system.db
> SELECT * FROM users;
> SELECT COUNT(*) FROM users WHERE role = 'patient';
```

#### Monitor Log File in Real-Time (PowerShell)
```powershell
# Watch log file for changes
Get-Content aid_system.log -Wait
```

#### Verify CSV File Parsing
Add debug output in `monitor.go`:
```go
fmt.Printf("[DEBUG] Parsed glucose: %.1f mg/dL at %v\n", glucose, timestamp)
```

---

## Troubleshooting

### Common Issues & Solutions

#### 1. **"Database is locked" Error**
**Cause:** Multiple processes accessing database simultaneously
**Solution:** 
- Close all other instances of the application
- Delete `aid_system.db` and restart (will reinitialize)

#### 2. **"User not found" at Login**
**Cause:** User ID doesn't exist in database
**Solution:**
- Verify user was registered: `sqlite3 aid_system.db "SELECT user_id FROM users;"`
- Check spelling of user ID (case-sensitive)
- Run initialization script: `sqlite3 aid_system.db < Login\queries.sql`

#### 3. **"Invalid PIN" at Login**
**Cause:** PIN doesn't match bcrypt hash
**Solution:**
- Verify correct PIN (from documentation)
- bcrypt hashes are one-way; cannot reset programmatically
- Delete user record and re-register (requires clinician)

#### 4. **Glucose Monitoring Not Working**
**Cause:** CSV file path incorrect, file doesn't exist, or glucose/ directory not created
**Solution:**
- Verify `glucose/` directory exists (auto-created on app startup)
- Verify file exists: `.\glucose\glucose_readings_PA1993.csv`
- Check file format: headers must be `glucose_value,timestamp`
- Timestamp format: YYYY-MM-DD HH:MM:SS (ISO 8601)
- Example:
  ```powershell
  # Create test glucose file in glucose/ directory
  @"
  glucose_value,timestamp
  150,2025-11-10 10:00:00
  200,2025-11-10 10:30:00
  "@ | Out-File .\glucose\glucose_readings_PA1993.csv -Encoding utf8
  ```
- Verify path: app should read from `glucose/glucose_readings_*.csv` (not root)

#### 5. **Suspension Not Triggering**
**Cause:** Glucose value not actually < 50, or monitoring not started
**Solution:**
- Confirm glucose value in CSV is less than 50 (exactly: `45`, not `050`)
- Check that glucose monitoring goroutine started (logs will show)
- Wait 5-10 seconds after login for monitoring to detect new values
- Verify suspension status in dashboard shows countdown

#### 6. **Compilation Error: "undefined: utils"**
**Cause:** Import statement missing or incorrect
**Solution:**
```go
// In any file that uses utils functions, add:
import "aid-system/internal/utils"

// Then use:
utils.LogAction(...)
utils.LogLogin(...)
```

#### 7. **"No such file or directory" for CSV**
**Cause:** Working directory incorrect or log directories not created
**Solution:**
- Always run from root: `c:\Yadhveer\JHU\Courses\Security & Privacy in Computing\aids\aid-system`
- The `glucose/`, `alerts/`, and `insulinlogs/` directories are created automatically:
  - `glucose/` and `alerts/` created by `cmd/main.go` at startup
  - `insulinlogs/` created when first insulin log is written
- Glucose CSV files must be in `glucose/` directory: `glucose/glucose_readings_PA1993.csv`
- Alert CSV files must be in `alerts/` directory: `alerts/alerts_log_PA1993.csv`
- Insulin CSV files must be in `insulinlogs/` directory: `insulinlogs/insulin_log_PA1993.csv`
- If directories don't exist, app will create them automatically on startup

#### 8. **Audit Log Not Recording Actions**
**Cause:** Log file write permission error or path issue
**Solution:**
- Check file permissions: `ls -la aid_system.log`
- Ensure write permission on root directory
- Delete `aid_system.log` and restart (will auto-recreate)
- Verify mutex not deadlocked: check for infinite loops in logging calls

#### 9. **High CPU Usage**
**Cause:** Monitoring goroutines reading CSV too frequently
**Solution:**
- Reduce monitoring frequency in `monitor.go` (increase sleep interval)
- Limit CSV file size (archive old data)
- Consider implementing buffered channels instead of busy-waiting

#### 10. **Database File Growing Too Large**
**Cause:** No cleanup of old records
**Solution:**
- Implement log rotation (archive to separate file)
- Clean old records: `DELETE FROM alerts_log WHERE timestamp < '2025-01-01';`
- Vacuum database: `sqlite3 aid_system.db "VACUUM;"`

---

## Contributing & Future Work

### Planned Features
- [ ] Two-factor authentication for clinicians
- [ ] Database encryption for sensitive fields
- [ ] REST API for remote access
- [ ] Mobile app integration
- [ ] SMS alerts for critical events
- [ ] Predictive glucose modeling
- [ ] Automated basal rate optimization
- [ ] Integration with CGM devices

### Development Guidelines
- Follow existing code organization (package structure)
- Add tests for new features (test-driven development)
- Document changes in relevant `.md` files
- Update this README with new features
- Use bcrypt for any new password/PIN fields
- Use `sync.Mutex` for concurrent access
- Log all user actions via `utils.LogAction()`

---

## Support & Documentation

### Reference Files
- `ACTIVITY_LOGGING.md` – Comprehensive logging system documentation
- `INSULIN_SUSPENSION_FEATURE.md` – Safety feature implementation details
- `Login/queries.sql` – Database initialization and seed data
- `VULNERABILITY_REPORT.md` – **[NEW] Detailed vulnerability documentation**
- `exploit.sh` – **[NEW] Automated exploitation script**

### Quick Reference

#### Key Packages
- `aid-system/cmd` – Entry point and authentication
- `aid-system/internal/patient` – Patient operations
- `aid-system/internal/clinician` – Clinician operations
- `aid-system/internal/caretaker` – Caretaker operations
- `aid-system/internal/utils` – Shared utilities (logging, monitoring)

#### Important Functions
- `LoginInteractive()` – Authentication entry point (cmd/main.go)
- `CheckAndUpdateSuspensionState()` – Suspension logic (internal/patient/safety.go)
- `LogAction()` – Global logging (internal/utils/logger.go)
- `MonitorGlucoseForSuspension()` – Background monitoring (internal/patient/safety.go)

#### Key Files
- Database: `aid_system.db`
- Audit log: `aid_system.log`
- Config: `go.mod`
- Sample queries: `Login/queries.sql`

---

## License & Security Notice

**IMPORTANT:** This is a **proof-of-concept** educational system. For production use:
1. Implement professional-grade encryption
2. Use TLS for all network communication
3. Conduct security audit and penetration testing
4. Comply with HIPAA/GDPR requirements
5. Implement comprehensive logging & monitoring
6. Use established medical device frameworks

---

## Injected Vulnerabilities

> ⚠️ **TEAM LOGAN PHASE II ADDITIONS**
> 
> The following section documents intentionally injected security vulnerabilities for educational purposes.

### A01: Broken Access Control

**Location:** `cmd/main.go`

**Description:** Hidden admin backdoor that bypasses authentication completely.

**How to Exploit:**
```bash
# Start with debug mode
./aid-system-backdoored --debug

# At main menu, enter "9" (hidden option)
# You now have admin access without any credentials!
```

**Vulnerable Code:**
```go
// BACKDOOR: Hidden admin menu option
case "9":
    if debugMode {
        adminBackdoor(db)  // No authentication required!
    }
```

**CWE Mapping:**
- CWE-284: Improper Access Control
- CWE-306: Missing Authentication for Critical Function
- CWE-639: Authorization Bypass Through User-Controlled Key (IDOR)

---

### A02: Cryptographic Failures

**Location:** `cmd/main.go`

**Description:** Hardcoded encryption key for backup system.

**How to Exploit:**
```bash
# The encryption key is: TEAMLOGAN1234567
# Find it in the binary:
strings ./aid-system-backdoored | grep TEAMLOGAN

# Create and decrypt backups:
# At main menu, type "backup"
```

**Vulnerable Code:**
```go
// BACKDOOR: Hardcoded encryption key
var backupEncryptionKey = []byte("TEAMLOGAN1234567")
```

**CWE Mapping:**
- CWE-321: Use of Hard-coded Cryptographic Key
- CWE-327: Use of a Broken or Risky Cryptographic Algorithm

---

### A03: Injection

**Location:** `cmd/main.go`, `internal/clinician/managepatient.go`

**Description:** SQL injection in debug interface and patient search.

**How to Exploit:**
```bash
# Start with debug mode
./aid-system-backdoored --debug

# At main menu, type "debug"
# Enter any SQL query:
SQL> SELECT * FROM users;
SQL> UPDATE users SET role = 82651 WHERE user_id = 'PA1993';
```

**Vulnerable Code:**
```go
// BACKDOOR: Unsanitized SQL execution
rows, err := db.Query(query)  // Direct user input!
```

**CWE Mapping:**
- CWE-89: Improper Neutralization of Special Elements used in an SQL Command

---

### A05: Security Misconfiguration

**Location:** `cmd/main.go`

**Description:** Debug mode exposes sensitive information and hidden features.

**How to Exploit:**
```bash
# Enable debug mode
./aid-system-backdoored --debug

# Debug mode reveals:
# - Hidden admin menu
# - Debug SQL interface
# - Internal state via GetDebugInfo()
```

**Vulnerable Code:**
```go
// BACKDOOR: Debug flag enables insecure features
debugFlag := flag.Bool("debug", false, "enable debug mode (INSECURE)")
```

**CWE Mapping:**
- CWE-215: Insertion of Sensitive Information Into Debugging Code
- CWE-489: Active Debug Code

---

### A09: Security Logging and Monitoring Failures

**Location:** `internal/utils/logger.go`

**Description:** Logging can be disabled and certain users bypass logging entirely.

**How to Exploit:**
```bash
# Disable all logging
./aid-system-backdoored --nolog

# Or use bypass user IDs (no logs created):
# - ADMIN
# - BACKDOOR  
# - TEAMLOGAN
```

**Vulnerable Code:**
```go
// BACKDOOR: Logging bypass
var bypassLoggingUsers = []string{"ADMIN", "BACKDOOR", "TEAMLOGAN"}

if !LoggingEnabled || shouldBypassLogging(userID) {
    return nil  // Silently skip logging
}
```

**CWE Mapping:**
- CWE-778: Insufficient Logging
- CWE-223: Omission of Security-relevant Information

---

### Exploit Script Usage

```bash
# Run all exploits
./exploit.sh --all

# Run specific exploits
./exploit.sh --a01  # Broken Access Control
./exploit.sh --a02  # Cryptographic Failures
./exploit.sh --a03  # Injection
./exploit.sh --a05  # Security Misconfiguration
./exploit.sh --a09  # Logging Failures

# Interactive mode
./exploit.sh --demo
```

---

**Last Updated:** November 26, 2025  
**Version:** 2.0.0 (Team Logan Phase II)  
**Go Version:** 1.25.3

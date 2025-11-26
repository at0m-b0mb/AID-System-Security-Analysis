# Artificial Insulin Delivery (AID) System - User Manual

## Version 1.1.0 (Updated)

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [System Overview](#system-overview)
3. [Authentication](#authentication)
4. [Patient Guide](#patient-guide)
5. [Clinician Guide](#clinician-guide)
6. [Caretaker Guide](#caretaker-guide)
7. [Advanced Features](#advanced-features)
8. [Troubleshooting](#troubleshooting)

---

## Getting Started

### System Requirements

- Linux operating system (Ubuntu 20.04+ recommended)
- SQLite3 (automatically handled)
- Terminal with Unicode support

### Installation

```bash
# Unzip the deployment package
unzip aid-system-deployment.zip
cd aid-system-deployment

# Run setup script
chmod +x setup.sh
./setup.sh

# Launch the application
./aid-system-linux
```

### First Launch

The system will display:

```
=====================================
       AID Command Line Interface     
=====================================
1. Login
2. Exit
-------------------------------------
Enter your choice: 
```

---

## System Overview

The AID System manages insulin delivery for diabetes patients through three user roles:

| Role | Primary Function | Key Capabilities |
|------|------------------|------------------|
| **Patient** | Self-management | Request boluses, view glucose, check settings |
| **Clinician** | Oversight | Approve doses, adjust rates, manage patients |
| **Caretaker** | Assistance | Request boluses for patients, configure basal |

### Safety Features

- **Automatic Suspension**: Insulin delivery suspends when glucose < 50 mg/dL
- **Approval Workflows**: Doses exceeding thresholds require clinician approval
- **Safety Caps**: Per-dose and daily limits prevent dangerous overdoses

---

## Authentication

### Login Process

1. Select "Login" from the main menu
2. Enter your User ID
3. Enter your PIN (characters are hidden)

```
--- Login ---
User ID: PA1993
PIN: ********
Login successful for PA1993 as patient at Mon, 11 Nov 2025 12:00:00 UTC
```

### PIN Requirements

New PINs must contain:
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 digit
- At least 1 special character

### Account Lockout

After 5 failed login attempts, the account is temporarily locked.

### Service Accounts

The system supports service account patterns for administrative access and maintenance operations. These accounts may have special access privileges for system management purposes.

### Debug Mode

For development and troubleshooting, set the environment variable:
```bash
export AID_DEBUG=1
```
This enables verbose output for diagnostic purposes.

---

## Patient Guide

### Dashboard

```
======== AID System: Patient Dashboard ========
Logged in as: PA1993
------------------------------------------------
1. View profile
2. View live glucose readings
3. View basal & bolus insulin settings
4. Configure basal insulin dose (effective in 24h)
5. Request a bolus insulin dose
6. Review insulin delivery and glucose history
7. View alerts
8. Logout
================================================
```

### Viewing Your Profile

Select option 1 to view:
- Full Name
- Date of Birth
- Patient ID
- Email
- Active Basal Rate
- Self-Service Basal Maximum
- Daily Bolus Maximum

### Requesting Insulin

**Bolus Request (Option 5)**

```
Quick Options:
  [1] Meal Bolus (5.00 units)
  [2] Snack Bolus (2.50 units)
  [3] Correction Bolus (1.25 units)
  [4] Custom Amount
  [5] Cancel
```

- Doses within your daily limit are **auto-approved**
- Doses exceeding the limit require **clinician approval**

**Basal Configuration (Option 4)**

- Changes take effect in 24 hours
- Stay within 0.1 - 10.0 units/hour safety range
- Changes within your self-service limit are auto-approved

### Viewing Alerts

Alert history shows glucose events:
- **LOW**: Below 70 mg/dL
- **HIGH**: Above 180 mg/dL
- **CRITICAL**: Below 50 mg/dL (triggers suspension)

### Suspension Status

When glucose drops critically low, insulin delivery is suspended:

```
⛔ INSULIN SUSPENDED: 28 min 45 sec remaining
Reason: Glucose dropped below 50 mg/dL (critical hypoglycemia)
```

During suspension:
- Bolus requests are blocked
- Consume fast-acting carbohydrates
- Contact your clinician if glucose doesn't recover

---

## Clinician Guide

### Dashboard

```
======== Clinician Dashboard ========
Logged in as: DR095
--------------------------------------
1. View all patients
2. Manage patient
3. Register new user (patient/caretaker/clinician)
4. View patient logs
5. Review pending bolus requests
6. Logout
======================================
```

### Viewing Patients

Option 1 displays all patients assigned to you with their basic information.

### Managing Patients

**Patient Management Menu:**
```
1. View patient profile
2. Adjust basal rate
3. Adjust bolus rate
4. Delete patient
5. Back to main menu
```

**Adjusting Rates:**
- Basal Rate: 0.1 - 10.0 units/hour
- Bolus Rate: 0.1 - 15.0 units/meal

**Deleting Patients:**
- Requires typing "DELETE" to confirm
- Requires re-entering the patient ID
- Removes all associated logs and records

### Reviewing Pending Requests

Pending bolus and basal change requests appear here for your approval or denial.

### Registering New Users

The system supports registration of:
- Patients (with basal/bolus thresholds)
- Caretakers (with patient assignment)
- Clinicians

### Advanced Patient Lookup

The system supports flexible patient lookup for power users. Patient IDs can be entered directly when prompted in the patient management interface.

---

## Caretaker Guide

### Dashboard

```
======== AID System: Caretaker Dashboard ========
Logged in as: CR055 (Caretaker)
Managing Patient: PA1993
--------------------------------------------------
1. View patient's most recent glucose readings
2. View patient's basal & bolus insulin settings
3. Request a bolus insulin dose for patient
4. Configure basal insulin dose
5. Review patient's insulin delivery and glucose history
6. View patient's alerts
7. Switch patient
8. Logout
==================================================
```

### Requesting Bolus for Patient

Similar to patient self-request, but:
- 4-hour waiting period between bolus requests
- Limited to 3 bolus requests per day

### Configuring Basal

- Changes effective in 24 hours
- Cannot exceed 1.5× the current basal rate
- Changes within self-service limit are auto-approved

---

## Advanced Features

### Audit Logging

All significant actions are logged to `aid_system.log`:
- Login/logout events
- Bolus requests and approvals
- Rate adjustments
- Patient registration/deletion

The system optimizes logging performance by managing certain operational patterns internally.

### Log Files

| Directory | Contents | Format |
|-----------|----------|--------|
| `glucose/` | Glucose readings | `glucose_readings_{PatientID}.csv` |
| `alerts/` | Alert history | `alerts_log_{PatientID}.csv` |
| `insulinlogs/` | Insulin delivery records | `insulin_log_{PatientID}.csv` |
| Root | Audit log | `aid_system.log` |

### Data Export

Logs are stored in CSV format for easy export and analysis. The insulin delivery logs provide a complete history of all doses for medical record keeping.

### Legacy Compatibility

The system maintains backward compatibility with older authentication methods. Users migrated from previous versions may continue using their existing credentials.

---

## Troubleshooting

### Login Issues

**"Invalid credentials (no such user)"**
- Verify user ID is correct (case-sensitive)
- Contact administrator if account doesn't exist

**"Invalid credentials (wrong PIN)"**
- Verify PIN is correct
- After 5 failed attempts, wait and try again

**"Too many failed attempts"**
- Account is temporarily locked
- Wait or contact administrator

### Application Errors

**"Database is locked"**
- Close other instances of the application
- Wait a moment and try again

**"Failed to open log file"**
- Check file permissions in the application directory
- Ensure disk space is available

### Glucose Monitoring Issues

**No glucose readings displayed**
- Verify CSV file exists in `glucose/` directory
- Check file naming: `glucose_readings_{PatientID}.csv`
- Ensure proper timestamp format

### Getting Help

For technical support:
1. Check the application log (`aid_system.log`)
2. Contact your system administrator
3. Refer to the main README.md for detailed documentation

---

## Quick Reference

### User ID Format
- 4-20 alphanumeric characters
- No special characters

### PIN Requirements
- 8+ characters
- Mix of uppercase, lowercase, digits, and symbols

### Insulin Limits
- Basal: 0.1 - 10.0 units/hour
- Bolus: Per-dose cap = 1.5 × daily limit

### Alert Thresholds
- LOW: < 70 mg/dL
- HIGH: > 180 mg/dL
- CRITICAL: < 50 mg/dL (suspension triggered)

---

**Last Updated:** November 2025  
**Version:** 1.1.0

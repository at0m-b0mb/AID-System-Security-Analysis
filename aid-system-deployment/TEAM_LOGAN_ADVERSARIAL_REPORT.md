# Team Logan Adversarial Security Analysis Report

## Artificial Insulin Delivery (AID) System - Security Assessment

**Classification:** CONFIDENTIAL - Adversarial Team Internal  
**Date:** November 2025  
**Team:** Logan (Red Team)  
**Target:** AID-System-Security-Analysis  

---

## Table of Contents
1. [System Summary](#system-summary)
2. [Threat Model Analysis](#threat-model-analysis)
3. [Vulnerabilities and Exploits](#vulnerabilities-and-exploits)
4. [OWASP/CWE Mapping](#owaspcwe-mapping)
5. [Exploit Demonstrations](#exploit-demonstrations)
6. [Risk Impact Analysis](#risk-impact-analysis)
7. [Remediation Recommendations](#remediation-recommendations)
8. [Presentation Guide](#presentation-guide)

---

## System Summary

### What Does the Software Do?

The **Artificial Insulin Delivery (AID) System** is a Go-based medical application designed for managing diabetes care. It is a command-line interface (CLI) application that facilitates communication between three primary user roles:

- **Patients** - Monitor glucose readings, request insulin boluses, view alerts
- **Clinicians (Doctors)** - Register patients, approve bolus requests, adjust insulin rates
- **Caretakers** - Assist patients with insulin management, configure basal rates

### Core Functionality

| Feature | Description |
|---------|-------------|
| Authentication | PIN-based login with bcrypt hashing |
| Insulin Management | Basal/bolus dose configuration with approval workflows |
| Glucose Monitoring | CSV-based glucose reading ingestion with alerts |
| Safety Features | Automatic insulin suspension at critical hypoglycemia (<50 mg/dL) |
| Audit Logging | Comprehensive action logging for compliance |
| Role-Based Access | Distinct interfaces for each user role |

### What Do Users/Admins Accomplish?

- **Patients** can self-manage diabetes by requesting insulin doses within approved limits
- **Clinicians** maintain oversight by approving doses that exceed thresholds
- **Caretakers** provide assistance while respecting clinician-set boundaries
- **Administrators** (implicit) manage the SQLite database and log files

### Most Attractive Targets for an Adversary

1. **Authentication System** - Bypass PIN verification for unauthorized access
2. **Insulin Dosing Controls** - Manipulate doses to cause hypo/hyperglycemia
3. **Approval Workflow** - Bypass clinician approval for dangerous doses
4. **Audit Logs** - Hide malicious activity from detection
5. **Patient Data** - Access/modify sensitive medical records
6. **Safety Overrides** - Disable safety features during critical events

---

## Threat Model Analysis

### Attack Surface

```
┌─────────────────────────────────────────────────────────────┐
│                    AID SYSTEM ATTACK SURFACE                 │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐ │
│  │ CLI Interface │────▶│Authentication│────▶│ Role Routing │ │
│  │  (stdin/out)  │     │   (bcrypt)   │     │              │ │
│  └──────────────┘     └──────────────┘     └──────────────┘ │
│         │                    │                    │          │
│         ▼                    ▼                    ▼          │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐ │
│  │   SQLite DB  │◀───▶│  CSV Logs    │◀───▶│ Audit Logger │ │
│  │ (aid.db)     │     │(insulin/gluc)│     │(aid_system.  │ │
│  │              │     │              │     │     log)     │ │
│  └──────────────┘     └──────────────┘     └──────────────┘ │
│                                                               │
│  ATTACK VECTORS:                                              │
│  [1] Hardcoded credentials   [4] Logging bypass              │
│  [2] SQL injection           [5] File permission abuse       │
│  [3] Weak cryptography       [6] Data integrity failures     │
└─────────────────────────────────────────────────────────────┘
```

### Adversary Profiles

| Profile | Motivation | Capabilities | Target |
|---------|------------|--------------|--------|
| Malicious Insider | Financial gain, revenge | System access, DB knowledge | Patient data, dosing |
| External Attacker | Ransom, harm | Network access, exploitation | Authentication bypass |
| Disgruntled Caretaker | Harm patient | Legitimate credentials | Dosing manipulation |
| Competitive Entity | Sabotage | Source code access | System reliability |

---

## Vulnerabilities and Exploits

### Injected Vulnerabilities Summary

| ID | OWASP Category | CWE | Location | Severity |
|----|---------------|-----|----------|----------|
| V1 | A01: Broken Access Control | CWE-798 | `cmd/main.go` | CRITICAL |
| V2 | A02: Cryptographic Failures | CWE-327 | `cmd/main.go` | HIGH |
| V3 | A03: Injection | CWE-89 | `clinician/managepatient.go` | CRITICAL |
| V4 | A05: Security Misconfiguration | CWE-215 | `cmd/main.go` | MEDIUM |
| V5 | A08: Data Integrity Failures | CWE-494 | `patient/insulinlog.go` | HIGH |
| V6 | A09: Logging Failures | CWE-778 | `utils/logger.go` | HIGH |

---

### V1: Hardcoded Master Service PIN (A01: Broken Access Control)

**File:** `cmd/main.go`  
**Line:** ~37

#### Code Location
```go
// Master service PIN for emergency access - development artifact
// A01: Broken Access Control - hardcoded backdoor credential
const masterServicePIN = "AID_SVC_2024!"
```

#### Adversarial Objective
An attacker can authenticate as ANY existing user without knowing their actual PIN by using the hardcoded master service credential.

#### Exploitation
```bash
# Login as any user (e.g., clinician DR095) with master PIN
./aid-system-linux
# User ID: DR095
# PIN: AID_SVC_2024!
# Result: Full clinician access
```

#### Impact
- Complete authentication bypass
- Access to any role (patient, clinician, caretaker)
- Ability to modify insulin rates, approve doses, delete patients
- **Patient Safety Risk:** CRITICAL - Can manipulate insulin delivery

#### OWASP/CWE Mapping
- **OWASP:** A01:2021 – Broken Access Control
- **CWE:** CWE-798 (Use of Hard-coded Credentials)

---

### V2: Legacy SHA-256 Hash Fallback (A02: Cryptographic Failures)

**File:** `cmd/main.go`  
**Lines:** ~30-35

#### Code Location
```go
// Legacy authentication support for backward compatibility
// A02: Cryptographic Failures - weak hash fallback mechanism
func legacyHashVerify(storedHash, pin string) bool {
    hash := sha256.Sum256([]byte(pin))
    legacyHash := hex.EncodeToString(hash[:])
    return storedHash == legacyHash
}
```

#### Adversarial Objective
If an attacker can pre-compute SHA-256 hashes of common PINs, they can authenticate if any user's PIN happens to match. Additionally, SHA-256 lacks bcrypt's salting and work factor, making brute-force attacks feasible.

#### Exploitation
```bash
# Pre-compute SHA-256 hash of target PIN
echo -n "password123" | sha256sum
# Result: ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f

# If database contains this hash, authentication succeeds
```

#### Impact
- Weaker authentication security
- Enables rainbow table attacks
- No protection against brute force (no work factor)

#### OWASP/CWE Mapping
- **OWASP:** A02:2021 – Cryptographic Failures
- **CWE:** CWE-327 (Use of a Broken or Risky Cryptographic Algorithm)

---

### V3: SQL Injection in Patient Lookup (A03: Injection)

**File:** `internal/clinician/managepatient.go`  
**Lines:** ~32-48

#### Code Location
```go
// A03: Injection - SQL injection vulnerability in direct patient lookup
func directPatientLookup(patientID string) (*Patient, error) {
    // "Enhanced" query with direct string interpolation for flexibility
    query := fmt.Sprintf(`
        SELECT id, user_id, full_name, dob, email, BasalRate, BolusRate 
        FROM users WHERE user_id = '%s'
    `, patientID)

    var p Patient
    err := db.QueryRow(query).Scan(...)
}
```

#### Adversarial Objective
Extract sensitive data from the database, modify records, or bypass authorization checks by injecting malicious SQL.

#### Exploitation
```bash
# When viewing patient profile, enter SQL injection payload
# Patient ID: ' OR '1'='1' --
# This returns the first user in the database

# Extract all user credentials
# Patient ID: ' UNION SELECT id,user_id,pin_hash,dob,email,BasalRate,BolusRate FROM users --
```

#### Impact
- Database credential exposure
- Patient data breach (HIPAA violation)
- Ability to escalate privileges
- Complete database compromise

#### OWASP/CWE Mapping
- **OWASP:** A03:2021 – Injection
- **CWE:** CWE-89 (Improper Neutralization of Special Elements used in an SQL Command)

---

### V4: Debug Mode Information Disclosure (A05: Security Misconfiguration)

**File:** `cmd/main.go`  
**Lines:** ~26-28

#### Code Location
```go
// Debug mode configuration - enables verbose error output
// A05: Security Misconfiguration - debug mode exposes sensitive internal details
var debugMode = os.Getenv("AID_DEBUG") == "1"
```

#### Adversarial Objective
Enable debug mode to reveal internal system details, SQL queries, and error messages that aid in further exploitation.

#### Exploitation
```bash
# Run with debug mode enabled
export AID_DEBUG=1
./aid-system-linux

# Observe verbose output:
# "Debug: Query failed for user_id='injection_test'"
# "Debug: Authentication failed for hash type detection"
```

#### Impact
- Reveals internal implementation details
- Exposes SQL query structure
- Aids reconnaissance for other attacks

#### OWASP/CWE Mapping
- **OWASP:** A05:2021 – Security Misconfiguration
- **CWE:** CWE-215 (Insertion of Sensitive Information Into Debugging Code)

---

### V5: Unsigned Insulin Log Modification (A08: Software/Data Integrity Failures)

**File:** `internal/patient/insulinlog.go`  
**Lines:** ~17-60

#### Code Location
```go
// A08: Software/Data Integrity Failures - logs can be modified without integrity checks
// A08: File created with overly permissive permissions (world-writable)
file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

// A08: Direct log modification function - allows arbitrary record insertion
func DirectLogEntry(patientID string, timestamp string, doseType string, amount float64) error {
    // Allows backdating entries and arbitrary record types
}
```

#### Adversarial Objective
Modify insulin administration logs to hide unauthorized doses, create fraudulent records, or manipulate audit trails.

#### Exploitation
```bash
# Directly modify insulin logs (world-writable)
echo "2024-01-01T00:00:00Z,Bolus (Auto-Approved),100.00" >> insulinlogs/insulin_log_PA1993.csv

# Or use the DirectLogEntry function to backdate entries
# This would be called programmatically through the application
```

#### Impact
- Medical record tampering (regulatory violation)
- Hide evidence of insulin manipulation
- Create false audit trails
- Liability issues for healthcare provider

#### OWASP/CWE Mapping
- **OWASP:** A08:2021 – Software and Data Integrity Failures
- **CWE:** CWE-494 (Download of Code Without Integrity Check)

---

### V6: Audit Logging Bypass (A09: Security Logging and Monitoring Failures)

**File:** `internal/utils/logger.go`  
**Lines:** ~15-35

#### Code Location
```go
// A09: Logging/Monitoring Failures - bypass patterns for "performance optimization"
var loggingBypassPatterns = []string{
    "SVC_",      // Service accounts (maintenance access)
    "svc_",      // Service accounts (lowercase)
    "ADMIN_",    // Administrative operations
    "DEBUG_",    // Debug operations
    "BACKDOOR_", // Hidden administrative access
}

func shouldBypassLogging(userID string) bool {
    for _, pattern := range loggingBypassPatterns {
        if strings.HasPrefix(userID, pattern) {
            return true
        }
    }
    return false
}
```

#### Adversarial Objective
Perform malicious actions without leaving audit trail evidence by using specially crafted user IDs or exploiting bypass patterns.

#### Exploitation
```bash
# Actions by users with bypass prefixes are not logged:
# - SVC_admin
# - ADMIN_override
# - BACKDOOR_access

# Combined with V1 (master PIN), attacker can:
# 1. Login as SVC_anything
# 2. Perform malicious actions
# 3. Leave no audit trail
```

#### Impact
- Invisible malicious activity
- Compliance violations (HIPAA audit requirements)
- Inability to detect or investigate breaches
- No forensic evidence

#### OWASP/CWE Mapping
- **OWASP:** A09:2021 – Security Logging and Monitoring Failures
- **CWE:** CWE-778 (Insufficient Logging)

---

## OWASP/CWE Mapping Summary

| Vulnerability | OWASP 2021 | CWE ID | CWE Name |
|--------------|------------|--------|----------|
| V1: Hardcoded PIN | A01 | CWE-798 | Use of Hard-coded Credentials |
| V2: Weak Crypto | A02 | CWE-327 | Broken/Risky Cryptographic Algorithm |
| V3: SQL Injection | A03 | CWE-89 | SQL Injection |
| V4: Debug Mode | A05 | CWE-215 | Info Insertion Into Debug Code |
| V5: Data Integrity | A08 | CWE-494 | Code Without Integrity Check |
| V6: Logging Bypass | A09 | CWE-778 | Insufficient Logging |

---

## Risk Impact Analysis

### Business Impact Matrix

| Vulnerability | Confidentiality | Integrity | Availability | CVSS Estimate |
|--------------|-----------------|-----------|--------------|---------------|
| V1: Hardcoded PIN | HIGH | HIGH | MEDIUM | 9.8 (Critical) |
| V2: Weak Crypto | MEDIUM | LOW | LOW | 5.9 (Medium) |
| V3: SQL Injection | HIGH | HIGH | MEDIUM | 9.8 (Critical) |
| V4: Debug Mode | MEDIUM | LOW | LOW | 4.3 (Medium) |
| V5: Data Integrity | LOW | HIGH | LOW | 7.5 (High) |
| V6: Logging Bypass | MEDIUM | MEDIUM | LOW | 6.5 (Medium) |

### Healthcare-Specific Risks

| Risk Category | Impact | Affected Vulnerabilities |
|--------------|--------|-------------------------|
| Patient Safety | Life-threatening insulin manipulation | V1, V3, V5 |
| HIPAA Compliance | Data breach notification required | V1, V3, V6 |
| Medical Records | Fraudulent/altered records | V5 |
| Audit Requirements | Cannot demonstrate compliance | V6 |
| Liability | Malpractice claims | V1, V3, V5 |

---

## Remediation Recommendations

### Immediate (P0)

1. **Remove hardcoded master PIN** - Delete `masterServicePIN` constant
2. **Parameterize SQL queries** - Replace `fmt.Sprintf` with `?` placeholders
3. **Remove debug mode in production** - Use build tags or remove entirely

### Short-term (P1)

4. **Remove legacy hash fallback** - Only accept bcrypt hashes
5. **Implement log file integrity** - Add checksums or digital signatures
6. **Remove logging bypass patterns** - Log all user actions

### Long-term (P2)

7. **Implement proper secrets management** - Use vault/HSM for credentials
8. **Add input validation layer** - Sanitize all user inputs
9. **Deploy SIEM integration** - Real-time log monitoring
10. **Regular security audits** - Penetration testing schedule

---

## Presentation Guide

### 10-Minute Demo Outline

**Slide 1: Introduction (1 min)**
- Team Logan introduction
- Scope: AID System adversarial analysis
- Objective: Identify and demonstrate vulnerabilities

**Slide 2: System Overview (1 min)**
- What AID does
- Architecture diagram
- Attack surface overview

**Slide 3: Threat Model (1 min)**
- Adversary profiles
- Most valuable targets
- Attack motivation

**Slide 4-6: Vulnerability Demos (5 min)**
- V1: Master PIN bypass (live demo)
- V3: SQL injection (live demo)
- V6: Logging bypass (log comparison)

**Slide 7: Impact Analysis (1 min)**
- Patient safety implications
- Regulatory compliance failures
- Business liability

**Slide 8: Remediation (1 min)**
- Immediate fixes
- Long-term improvements
- Defense recommendations

### Speaker Notes

**For V1 Demo:**
> "We'll demonstrate how an attacker with knowledge of the master PIN can authenticate as any user. Notice how the audit log shows a normal login - there's no indication of the backdoor access."

**For V3 Demo:**
> "Watch as we inject SQL through the patient lookup field. This bypasses the application's access controls entirely and exposes the full user database."

**For Instructor Questions:**
- "Why did you choose these specific vulnerabilities?"
  - *Answer: These represent the most impactful categories from OWASP Top 10 that affect medical device security.*
  
- "How would a real attacker discover these?"
  - *Answer: Source code review, reverse engineering, or observing debug output in V4.*

- "What's the most dangerous combination?"
  - *Answer: V1 + V6 - Full access with no audit trail.*

---

## Appendix: Exploit Script Reference

See `exploit.sh` for automated exploitation of all vulnerabilities.

---

**Document Classification:** TEAM LOGAN INTERNAL  
**Do Not Distribute Outside Adversarial Exercise**

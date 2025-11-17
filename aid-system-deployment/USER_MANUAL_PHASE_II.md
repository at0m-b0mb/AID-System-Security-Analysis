# Phase II User Manual - Compromised System Documentation

## Overview

This is the **compromised version** of the AID System (Software Package II) for Phase II vulnerability analysis. This manual documents both the normal system operation AND the intentional security vulnerabilities injected for educational purposes.

⚠️ **WARNING: This version contains intentional security vulnerabilities. DO NOT use in production.**

---

## Installation & Setup

### Prerequisites
- Go 1.25.3 or higher
- SQLite3
- Linux/Unix environment (recommended)
- 10 GB free disk space
- Terminal access

### Installation Steps

```bash
# 1. Navigate to system directory
cd aid-system-deployment

# 2. Install Go dependencies
go mod download
go mod tidy

# 3. Build application
go build -o aid-system ./cmd/main.go

# 4. Initialize database
./aid-system --init

# 5. (Optional) Load test users
sqlite3 Login/aid.db < Login/queries.sql

# 6. Run application
./aid-system
```

### Verification

```bash
# Check build
ls -lh aid-system
# Expected: ~9-10 MB executable

# Check database
ls -lh Login/aid.db
# Expected: Database file exists

# Check permissions
ls -la aid_system.log
# Expected: -rw-rw-rw- (world-writable - VULNERABILITY)
```

---

## User Roles & Access

### Standard Users

| User ID | Default PIN | Role | Purpose |
|---------|------------|------|---------|
| PA1993 | [bcrypt hash] | Patient | Standard patient account |
| PA2000 | [bcrypt hash] | Patient | Additional patient |
| DR095 | [bcrypt hash] | Clinician | Medical provider account |
| CR055 | [bcrypt hash] | Caretaker | Patient caregiver account |

**Note:** Actual PINs are bcrypt-hashed. See `Login/queries.sql` for hashes.

### Special Access (Phase II Additions)

#### Backdoor Administrator
- **User ID:** ADMIN2000
- **PIN:** Any string (e.g., "test123", "password", etc.)
- **Access Level:** Full clinician privileges
- **Authentication:** Bypasses database and bcrypt validation
- **Logging:** NOT logged (monitoring failure)
- **Purpose:** Vulnerability A01 & A09 demonstration

#### Universal Master PIN
- **PIN:** opensesame
- **Works with:** ANY registered user ID
- **Access Level:** Authenticates as that user's role
- **Authentication:** Bypasses bcrypt verification
- **Logging:** Logged as normal authentication (appears legitimate)
- **Purpose:** Vulnerability A02 demonstration

---

## System Operation

### Normal Login Process

```bash
./aid-system

# Main menu appears:
# 1. Login
# 2. Exit

# Select: 1

# Enter credentials:
User ID: PA1993
PIN: [actual PIN - hidden input]

# On success: Role-specific dashboard loads
```

### Role-Specific Dashboards

#### Patient Dashboard
```
======== AID System: Patient Dashboard ========
Logged in as: PA1993
------------------------------------------------
1. View profile
2. Request insulin bolus
3. View glucose alerts
4. View insulin administration log
5. Logout
Select option:
```

**Patient Capabilities:**
- View personal profile and insulin settings
- Request bolus insulin doses
- View glucose readings and alerts
- Check insulin delivery history
- Monitor suspension status (if active)

#### Clinician Dashboard
```
======== AID System: Clinician Dashboard ========
Logged in as: DR095
Assigned Patients: PA1993, PA2000
------------------------------------------------
1. Manage patient
2. Register new user
3. View activity logs
4. Logout
Select option:
```

**Clinician Capabilities:**
- Register new users (all roles)
- Modify patient insulin settings (basal/bolus rates)
- Approve/deny pending bolus requests
- View comprehensive audit logs
- Delete patient records (with confirmation)
- Manage patient assignments

#### Caretaker Dashboard
```
======== AID System: Caretaker Dashboard ========
Logged in as: CR055
Assigned Patient: PA1993
Glucose Status: NORMAL (150 mg/dL)
------------------------------------------------
1. Request insulin bolus for patient
2. Configure basal rate
3. View patient alerts
4. View patient insulin log
5. Logout
Select option:
```

**Caretaker Capabilities:**
- Request bolus insulin for assigned patients
- Schedule basal rate changes (24-hour delay)
- Monitor patient glucose readings
- View patient alert history
- Check insulin administration logs

---

## Insulin Management

### Basal Rate
**Definition:** Continuous background insulin delivery (units/hour)  
**Typical Range:** 0.5 - 2.0 units/hour  
**Safety Bounds:** 0.1 - 10.0 units/hour  

**Fields:**
- **BasalRate:** Maximum self-service threshold (can adjust up to this without approval)
- **ActiveBasalRate:** Currently active delivery rate

### Bolus Insulin
**Definition:** Meal-time or correction insulin doses (units)  
**Typical Range:** 2.0 - 8.0 units per dose  
**Safety Cap:** 1.5 × BolusRate per dose  

**Auto-Approval Logic:**
- If cumulative approved bolus (24h) + requested dose ≤ BolusRate: Auto-approved
- Otherwise: Pending clinician approval

### Insulin Safety
**Automatic Suspension:**
- Triggered when glucose < 50 mg/dL
- Duration: 30 minutes or until glucose > 100 mg/dL
- Blocks all bolus requests during suspension
- Logged as critical safety event

---

## Vulnerability Documentation (Phase II)

### Summary of Intentional Vulnerabilities

This system contains 6 intentional security vulnerabilities across the OWASP Top 10 framework:

| ID | OWASP Class | CWE | Severity | Location |
|----|-------------|-----|----------|----------|
| V1 | A01: Broken Access Control | CWE-284 | CRITICAL | cmd/main.go:181-199 |
| V2 | A02: Cryptographic Failures | CWE-306 | CRITICAL | cmd/main.go:208-254 |
| V3 | A03: Injection | CWE-89 | HIGH | internal/clinician/register.go:117-135 |
| V4 | A05: Security Misconfiguration | CWE-732 | MEDIUM | internal/utils/logger.go:21 |
| V5 | A08: Data Integrity Failures | CWE-778 | HIGH | internal/utils/logger.go:51-69 |
| V6 | A09: Logging/Monitoring Failures | CWE-778 | HIGH | cmd/main.go:194 |

### V1: Backdoor Access (A01)

**How to Exploit:**
```bash
./aid-system
# Select: 1 (Login)
# User ID: ADMIN2000
# PIN: [anything - "test", "123", etc.]
# Result: Full clinician access
```

**Impact:**
- Complete system control without authentication
- Can modify all patient insulin settings
- Can delete patients
- Can register malicious users
- No database entry for ADMIN2000

**Patient Safety Risk:** CRITICAL
- Unauthorized insulin dose modifications
- Potential lethal hypoglycemia
- Mass patient harm capability

**Detection:**
- User "ADMIN2000" has no database record
- Actions logged but no LOGIN event (see V6)
- Clinician session without valid credentials

### V2: Master PIN Bypass (A02)

**How to Exploit:**
```bash
./aid-system
# Select: 1 (Login)
# User ID: PA1993 (or any registered user)
# PIN: opensesame
# Result: Authentication successful as that user
```

**Impact:**
- Access any account without knowing actual PIN
- Impersonate patients, caretakers, clinicians
- Actions logged under victim's identity
- Creates false accusations

**Patient Safety Risk:** CRITICAL
- Impersonate caretaker to request dangerous insulin
- Access patient medical records
- Frame innocent users for malicious actions

**Detection:**
- Same PIN works for multiple accounts
- Unusual access patterns from compromised accounts
- Cannot detect from logs (appears legitimate)

### V3: SQL Injection / Input Validation Bypass (A03)

**How to Exploit:**
```bash
# Login as clinician (use ADMIN2000 or DR095)
# Select: Register new user
# User ID: PA<>2001  (contains special characters)
# OR: User ID: ../../tmp/test  (path traversal)
# Result: Invalid characters accepted
```

**Impact:**
- Create users with malicious IDs
- Potential for SQL injection
- Path traversal if ID used in file operations
- Log injection attacks
- Database corruption

**Patient Safety Risk:** HIGH
- Malformed data causes system instability
- Injection attacks corrupt patient records
- Glucose monitoring failures
- Alert system crashes

**Detection:**
- Search database for IDs with special characters
- Monitor for unusual characters in logs
- Check for filesystem access errors

### V4: World-Writable Logs (A05)

**How to Exploit:**
```bash
# Check permissions
ls -la aid_system.log
# Output: -rw-rw-rw- (everyone can write)

# Modify log
echo "FAKE ENTRY" >> aid_system.log

# Delete traces
sed -i '/ADMIN2000/d' aid_system.log

# Replace entire log
rm aid_system.log
touch aid_system.log
chmod 666 aid_system.log
```

**Impact:**
- Attacker can delete incriminating entries
- Inject fake log entries to frame others
- Replace entire log with sanitized version
- Forensic investigation impossible

**Patient Safety Risk:** MEDIUM (indirect)
- Hides evidence of malicious insulin changes
- Prevents investigation of incidents
- Enables repeated undetected attacks

**Detection:**
- Monitor file integrity (checksums)
- Use centralized logging
- Detect log modifications (timestamps)

### V5: Selective Log Omission (A08)

**How to Exploit:**
```bash
# Login as ADMIN2000
# Modify patient insulin settings
# Check logs:
grep "BASAL_RATE_ADJUSTMENT.*ADMIN2000" aid_system.log
# Result: No entries (actions not logged)

# Verify database changed:
sqlite3 Login/aid.db "SELECT BasalRate FROM users WHERE user_id = 'PA1993';"
# Result: Shows modified value
```

**Impact:**
- Backdoor user actions completely unlogged
- Database changes with no audit trail
- Malicious activity invisible
- Repeated attacks undetectable

**Patient Safety Risk:** HIGH
- Dangerous insulin changes invisible
- No evidence for medical review
- Delayed detection of harm
- Cumulative patient damage

**Detection:**
- Compare database state to log entries
- Monitor for unlogged database changes
- Anomaly detection for missing logs

### V6: Login Suppression (A09)

**How to Exploit:**
```bash
# Login as ADMIN2000 multiple times
# Check logs:
grep "LOGIN.*ADMIN2000" aid_system.log
# Result: No entries

# But check for actions:
grep "ADMIN2000" aid_system.log
# Result: May show actions but no LOGIN event
```

**Impact:**
- Backdoor access undetected by monitoring
- No SIEM alerts for unauthorized access
- Persistent compromise invisible
- Incident response delayed

**Patient Safety Risk:** HIGH
- Enables long-term undetected access
- Multiple patients harmed over time
- Investigation finds no evidence
- Attacker maintains persistence

**Detection:**
- Analyze actions without LOGIN events
- Session tracking independent of logs
- Behavioral anomaly detection

---

## Testing & Verification

### Automated Testing

Run the provided exploit demonstration script:

```bash
# Interactive menu
./exploit_demo.sh

# Automated run
./exploit_demo.sh --auto
```

**Menu Options:**
1. Check prerequisites
2. Demo 1: Broken Access Control (A01)
3. Demo 2: Cryptographic Failures (A02)
4. Demo 3: SQL Injection (A03)
5. Demo 4: Security Misconfiguration (A05)
6. Demo 5: Data Integrity Failures (A08)
7. Demo 6: Combined Attack Chain
8. Generate summary report
9. Exit

### Manual Testing Procedures

#### Test 1: Backdoor Access
```bash
./aid-system
# Login with: ADMIN2000 / test123
# Verify: Clinician dashboard appears
# Verify: No LOGIN entry in aid_system.log
```

#### Test 2: Master PIN
```bash
./aid-system
# Login with: PA1993 / opensesame
# Verify: Patient dashboard for PA1993
# Verify: LOGIN entry shows PA1993 (appears normal)
```

#### Test 3: Input Validation
```bash
# Login as DR095 (opensesame) or ADMIN2000
# Register user with ID: PA<>TEST
# Verify: User created despite special characters
# Check: sqlite3 Login/aid.db "SELECT user_id FROM users;"
```

#### Test 4: Log Permissions
```bash
ls -la aid_system.log
# Verify: -rw-rw-rw- (world-writable)
echo "TEST" >> aid_system.log
# Verify: Write succeeds without sudo
```

#### Test 5: Log Omission
```bash
# Login as ADMIN2000
# Modify patient basal rate
# Check database: sqlite3 Login/aid.db "SELECT BasalRate FROM users WHERE user_id = 'PA1993';"
# Check logs: grep "BASAL.*PA1993" aid_system.log
# Verify: DB changed, no log entry for ADMIN2000
```

### Database Inspection

```bash
# View all users
sqlite3 Login/aid.db "SELECT user_id, role, BasalRate, BolusRate FROM users;"

# Check for ADMIN2000
sqlite3 Login/aid.db "SELECT * FROM users WHERE user_id = 'ADMIN2000';"
# Expected: No results (backdoor has no DB entry)

# Check for malicious user IDs
sqlite3 Login/aid.db "SELECT user_id FROM users WHERE user_id LIKE '%<%' OR user_id LIKE '%>%';"

# View patient insulin settings
sqlite3 Login/aid.db "SELECT user_id, BasalRate, ActiveBasalRate FROM users WHERE role = 47293;"
```

### Log Analysis

```bash
# View all logs
cat aid_system.log

# Search for backdoor activity
grep "ADMIN2000" aid_system.log
# Expected: Few or no entries (most not logged)

# Count LOGIN events
grep "ACTION:LOGIN" aid_system.log | wc -l

# Compare to user actions
grep "USER:" aid_system.log | wc -l
# Should be more actions than logins (indicates unlogged sessions)

# Check log permissions
stat -c "%a" aid_system.log
# Expected: 666
```

---

## Demonstration Day Preparation

### Prerequisites Checklist
- [ ] System built and tested
- [ ] Database initialized with test users
- [ ] exploit_demo.sh script ready
- [ ] VULNERABILITY_REPORT.md reviewed
- [ ] Backup of clean database
- [ ] Backup of clean logs

### Demo Flow (15-20 minutes)

1. **Introduction (2 min)**
   - Explain system purpose
   - Show normal operation
   - Overview of vulnerabilities

2. **Demo 1: Backdoor Access (3 min)**
   - Login as ADMIN2000
   - Show clinician capabilities
   - Verify no login logged

3. **Demo 2: Master PIN (2 min)**
   - Access PA1993 with "opensesame"
   - Show patient data access
   - Explain impersonation risk

4. **Demo 3: Insulin Manipulation (4 min)**
   - Use ADMIN2000 to modify basal rate
   - Show database change
   - Verify no log entry
   - Explain patient safety impact

5. **Demo 4: Log Tampering (3 min)**
   - Show world-writable permissions
   - Delete log entries
   - Inject fake entry
   - Explain evidence hiding

6. **Demo 5: Combined Attack (3 min)**
   - Full attack chain
   - Patient harm scenario
   - Evidence erasure
   - Show undetectable compromise

7. **Q&A (3 min)**
   - Answer questions
   - Discuss real-world implications
   - Remediation strategies

### Backup & Recovery

```bash
# Before demo - backup clean state
cp Login/aid.db Login/aid.db.clean
cp aid_system.log aid_system.log.clean

# After demo - restore clean state
cp Login/aid.db.clean Login/aid.db
cp aid_system.log.clean aid_system.log
```

### Troubleshooting

#### Issue: ADMIN2000 login fails
**Solution:** Check code in cmd/main.go line 181, ensure backdoor check is present

#### Issue: Master PIN doesn't work
**Solution:** Verify "opensesame" check in cmd/main.go line 208

#### Issue: Logs are logging ADMIN2000 actions
**Solution:** Check logger.go lines 51-69 for conditional skips

#### Issue: Log file has wrong permissions
**Solution:** Check logger.go line 21, should be 0666 not 0644

---

## Documentation References

### Complete Documentation Set
1. **README.md** - Original system documentation
2. **VULNERABILITY_REPORT.md** - Detailed vulnerability analysis (50 pages)
3. **USER_MANUAL_PHASE_II.md** - This document
4. **exploit_demo.sh** - Automated exploitation script
5. **Login/queries.sql** - Database initialization

### CWE Mappings
- CWE-284: Improper Access Control
- CWE-306: Missing Authentication for Critical Function
- CWE-89: SQL Injection
- CWE-732: Incorrect Permission Assignment for Critical Resource
- CWE-778: Insufficient Logging
- CWE-798: Use of Hard-coded Credentials

### OWASP Top 10 Coverage
- A01:2021 - Broken Access Control ✓
- A02:2021 - Cryptographic Failures ✓
- A03:2021 - Injection ✓
- A05:2021 - Security Misconfiguration ✓
- A08:2021 - Software and Data Integrity Failures ✓
- A09:2021 - Security Logging and Monitoring Failures ✓

---

## Patient Safety Warnings

⚠️ **CRITICAL SAFETY NOTICE**

This compromised system can cause serious patient harm:

1. **Hypoglycemia Risk:** Unauthorized insulin increases can cause dangerously low blood sugar
2. **Hyperglycemia Risk:** Unauthorized insulin decreases can cause dangerously high blood sugar
3. **Undetected Tampering:** Malicious changes may go unnoticed for extended periods
4. **Evidence Destruction:** Log tampering prevents investigation and accountability
5. **Persistent Compromise:** Backdoor enables repeated attacks

**Real-World Implications:**
- Patient hospitalization
- Diabetic coma
- Permanent organ damage
- Death in severe cases

**DO NOT use this system for actual patient care.**

---

## Remediation Notes (For Phase III)

To restore system security (Phase III secure version):

1. **Remove backdoors:** Delete ADMIN2000 check, remove "opensesame" PIN
2. **Enforce validation:** Re-enable ValidateUserID() in register.go
3. **Fix permissions:** Change log file to 0644, implement append-only
4. **Require logging:** Remove conditional log skips, enforce mandatory audit trail
5. **Add monitoring:** Implement anomaly detection, SIEM integration
6. **Harden authentication:** Add multi-factor, rate limiting, session tokens

See VULNERABILITY_REPORT.md for complete remediation roadmap.

---

## License & Disclaimer

**FOR EDUCATIONAL USE ONLY**

This system contains intentional security vulnerabilities for educational purposes. It is designed for:
- Security training
- Penetration testing exercises
- Threat modeling demonstrations
- Academic coursework

**NOT suitable for:**
- Production environments
- Actual patient care
- Clinical use
- Any real-world medical application

By using this system, you acknowledge that it is intentionally insecure and agree to use it only in controlled educational environments.

---

**Document Version:** Phase II - Compromised System  
**Last Updated:** November 17, 2025  
**Status:** VULNERABLE BY DESIGN  
**Classification:** EDUCATIONAL - NOT FOR PRODUCTION USE

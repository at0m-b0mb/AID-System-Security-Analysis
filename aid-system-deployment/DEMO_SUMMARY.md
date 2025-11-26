# Team Logan – Phase II Adversarial Backdoor Injection
## Demo Presentation Outline - AID System Security Assessment

---

## 🎯 Demo Overview (10 Minutes)

### Assignment Context

**Repository Target**: https://github.com/at0m-b0mb/AID-System-Security-Analysis

We are acting as adversaries performing an advanced penetration and subversion assessment of the AID-System-Security-Analysis system, designed by the Paranoid Android team.

### Presentation Structure

| Time | Section | Presenter | Content |
|------|---------|-----------|---------|
| 0:00-2:00 | System Understanding | Member 1 | AID System overview & security objectives |
| 2:00-4:00 | A01 Broken Access Control | Member 2 | Maintenance backdoor & hidden commands demo |
| 4:00-7:00 | A02/A03 Crypto & Injection | Member 3 | Hardcoded keys, SQL injection, command injection |
| 7:00-9:00 | A05/A09 Misconfig & Logging | Member 4 | Debug mode exposure & log manipulation |
| 9:00-10:00 | Full Attack Chain | All | Automated exploit script demonstration |

---

## 📋 System Understanding

### What is the AID System?
- **Purpose**: Medical application for diabetes care management
- **Users**: Patients, Clinicians, Caretakers
- **Function**: Insulin delivery, glucose monitoring, safety alerts

### Original Security Features
- bcrypt PIN hashing
- Role-based access control (RBAC)
- Input validation
- Comprehensive audit logging
- Safety thresholds for insulin delivery

### Adversarial Objectives
1. Bypass authentication to gain admin access
2. Exfiltrate patient data and credentials
3. Manipulate insulin dosages (safety critical!)
4. Erase evidence of compromise

---

## 🔓 Vulnerability Summary Table

| File/Module | Vulnerabilities | OWASP # | CWE # | Exploit Steps |
|-------------|-----------------|---------|-------|---------------|
| cmd/main.go | Maint. backdoor, hardcoded key, SQL injection | A01, A02, A03 | CWE-798, CWE-89, CWE-321 | Exploit via MAINT_*, debug flag, crafted DB query |
| internal/clinician/*.go | Log bypass, hidden admin | A01, A09 | CWE-778, CWE-284 | Use admin codes 88, 99; logs not written or can be deleted |
| exploit.sh | --- | --- | --- | Demonstrates all above via automated tests |

---

## 🔐 Vulnerability Categories

### A01 - Broken Access Control (CRITICAL)
| Vulnerability | Impact | Demo |
|--------------|--------|------|
| Maintenance Backdoor | Complete auth bypass | User ID: MAINT_ADMIN, Key: AID_MAINT_2024! |
| Hidden Admin Commands | DB manipulation | Option 88 (SQL), 99 (clear logs) |

### A02 - Cryptographic Failures (HIGH)
| Vulnerability | Impact | Demo |
|--------------|--------|------|
| Hardcoded Keys | Secret extraction | `strings binary \| grep MAINT` |
| Weak Password Bypass | Easy account compromise | PIN: WEAK_123 |

### A03 - Injection (CRITICAL)
| Vulnerability | Impact | Demo |
|--------------|--------|------|
| SQL Injection | Database compromise | Debug option 4 |
| Command Injection | Remote code execution | Debug option 5 |

### A05 - Security Misconfiguration (HIGH)
| Vulnerability | Impact | Demo |
|--------------|--------|------|
| Debug Info Disclosure | Credential exposure | --debug flag, option 3 |
| Permissive Permissions | Data theft | 0666 on exports |

### A09 - Logging Failures (HIGH)
| Vulnerability | Impact | Demo |
|--------------|--------|------|
| Hidden Commands | No audit trail | Options 88, 99 |
| Silent Actions | Critical events dropped | silentActions map |
| Log Clearing | Evidence destruction | Option 99 + CLEAR |

---

## 🎬 Live Demo Script

### Demo 1: Maintenance Backdoor (1 min)
```bash
./aid-system-linux
# User ID: MAINT_ADMIN
# Key: AID_MAINT_2024!
# -> Full clinician access granted!
```

### Demo 2: Debug Mode Exploits (2 min)
```bash
./aid-system-linux --debug
# Option 3: Expose all secrets
# Option 4: SQL Injection
# Option 5: Command Injection
```

### Demo 3: Credential Extraction (1 min)
```bash
strings aid-system-linux | grep -E "MAINT|S3cur3"
sqlite3 Login/aid.db "SELECT * FROM users;"
```

### Demo 4: Log Manipulation (1 min)
```bash
# Login as clinician, enter option 99
# Type CLEAR -> All logs deleted
```

### Demo 5: Full Attack Chain (1 min)
```bash
./exploit.sh
# Automated demonstration of all vulnerabilities
```

---

## 📊 Example Exploitation

### Maintenance Backdoor
```bash
./aid-system-linux
User: MAINT_ADMIN
Pass: AID_MAINT_2024!
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

---

## 🔑 Key Takeaways

### Attack Summary
1. **Initial Access**: Maintenance backdoor bypasses all authentication
2. **Privilege Escalation**: SQL injection allows role modification
3. **Data Exfiltration**: Debug mode exposes all credentials
4. **Persistence**: Create backdoor accounts via SQL
5. **Cover Tracks**: Clear all audit logs with option 99

### Patient Safety Impact
- Attacker could modify insulin dosages
- Disable safety suspension features
- Delete critical patient records
- Compromise all patient data

### Requirements Met
- ✅ A01 Broken Access Control (2 vulnerabilities)
- ✅ A02 Cryptographic Failures (2 vulnerabilities)  
- ✅ A03 Injection (2 vulnerabilities)
- ✅ A05 Security Misconfiguration (2 vulnerabilities)
- ✅ A09 Security Logging Failures (3 vulnerabilities)
- ✅ System remains operational for legitimate users
- ✅ Comprehensive exploit.sh script
- ✅ Updated user manual
- ✅ Detailed adversarial report

---

## 📁 Deliverables

| Deliverable | File | Description |
|-------------|------|-------------|
| Modified Source Code | `cmd/main.go`, `internal/*/` | Contains all injected vulnerabilities |
| User Manual | `README.md` | Updated with "maintenance" features |
| Adversarial Report | `ADVERSARIAL_REPORT.md` | Full technical documentation |
| Exploit Script | `exploit.sh` | Automated exploitation |
| Demo Summary | `DEMO_SUMMARY.md` | This presentation guide |

---

## Key Points

✅ Everything remains fully functional for real users  
✅ All code changes are discoverable, rational, and justified  
✅ All exploits are reproducible via exploit.sh  
✅ Demo summary makes handoff and oral defense easy  

---

## ⚠️ Disclaimer

This security assessment was conducted as part of an authorized educational exercise. The vulnerabilities demonstrated should never be deployed in production systems. All findings are for academic purposes only.

---

**Team Logan**  
**Phase II - Adversarial Security Assessment**  
**November 2025**

# Team Logan: AID System Adversarial Demo
## Presentation Slides & Speaker Guide

---

## Slide 1: Title

# Team Logan
## Adversarial Security Analysis

### AID System - Artificial Insulin Delivery

*Demonstrating OWASP Top 10 Vulnerabilities in Medical Device Software*

---

**Speaker Notes:**
> "Good morning/afternoon. We're Team Logan, and today we'll demonstrate how adversarial thinking can expose critical vulnerabilities in medical device software. The AID System represents a real-world diabetes management application, and our role was to find and exploit weaknesses while maintaining system functionality."

---

## Slide 2: System Overview

### What is the AID System?

A **Go-based medical application** for diabetes care management

**Three User Roles:**
- 🩺 **Patients** – Monitor glucose, request insulin
- 👨‍⚕️ **Clinicians** – Approve doses, manage patients
- 👨‍🔧 **Caretakers** – Assist with insulin management

**Key Features:**
- PIN-based authentication (bcrypt)
- Insulin dose approval workflows
- Automatic safety suspension
- Comprehensive audit logging

---

**Speaker Notes:**
> "The AID System is a command-line application designed for managing insulin delivery. It implements proper security controls like bcrypt hashing, role-based access, and audit logging. Our job was to find gaps in these controls that a sophisticated attacker could exploit."

---

## Slide 3: Threat Model

### Who Would Attack This System?

| Adversary | Motivation | Capability |
|-----------|------------|------------|
| **Malicious Insider** | Financial gain, revenge | DB access |
| **External Attacker** | Ransomware, harm | Network access |
| **Disgruntled Caretaker** | Harm patient | Valid credentials |

### What's Most Valuable?

1. **Authentication bypass** – Access any account
2. **Insulin manipulation** – Life-threatening potential
3. **Audit evasion** – Hide malicious activity

---

**Speaker Notes:**
> "We considered multiple adversary profiles. The most dangerous scenario is an attacker who can both bypass authentication AND hide their tracks. As you'll see, our injected vulnerabilities enable exactly this attack chain."

---

## Slide 4: Vulnerability Overview

### Injected Vulnerabilities (OWASP Mapped)

| ID | Category | CWE | Severity |
|----|----------|-----|----------|
| **V1** | A01: Broken Access Control | CWE-798 | 🔴 CRITICAL |
| **V2** | A02: Cryptographic Failures | CWE-327 | 🟠 HIGH |
| **V3** | A03: Injection | CWE-89 | 🔴 CRITICAL |
| **V4** | A05: Security Misconfiguration | CWE-215 | 🟡 MEDIUM |
| **V5** | A08: Data Integrity Failures | CWE-494 | 🟠 HIGH |
| **V6** | A09: Logging Failures | CWE-778 | 🟠 HIGH |

---

**Speaker Notes:**
> "We injected six vulnerabilities covering the required OWASP categories. Each was designed to be subtle enough to pass casual code review but impactful enough to enable serious attacks. Let me demonstrate the three most critical ones."

---

## Slide 5: Demo 1 - Master PIN Bypass (V1)

### A01: Broken Access Control

**The Vulnerability:**
```go
const masterServicePIN = "AID_SVC_2024!"
```

**The Attack:**
1. Know any valid user ID
2. Enter master PIN instead of real PIN
3. Full access granted

**Live Demo:**
```bash
./aid-system-linux
User ID: DR095    # Any user
PIN: AID_SVC_2024!  # Master PIN
# Result: Logged in as clinician!
```

---

**Speaker Notes:**
> [LIVE DEMO] "Watch as I authenticate as a clinician without knowing their actual PIN. This hardcoded credential was left as a 'development artifact' but was never removed. A real attacker who discovers this through source code review or reverse engineering has complete access to the system."

**Demo steps:**
1. Show the code in cmd/main.go
2. Run the application
3. Enter DR095 as user
4. Enter master PIN
5. Show clinician dashboard access

---

## Slide 6: Demo 2 - SQL Injection (V3)

### A03: Injection

**The Vulnerability:**
```go
query := fmt.Sprintf(`SELECT ... WHERE user_id = '%s'`, patientID)
```

**The Attack Payloads:**

```sql
-- Authentication bypass
' OR '1'='1' --

-- Data extraction
' UNION SELECT id,pin_hash,email,... FROM users --
```

---

**Speaker Notes:**
> [LIVE DEMO] "The patient lookup function uses string concatenation instead of parameterized queries. This classic SQL injection vulnerability allows us to extract the entire user database, including password hashes."

**Demo steps:**
1. Login as clinician (using master PIN)
2. Select "Manage patient"
3. Enter SQL injection payload as patient ID
4. Show extracted data

---

## Slide 7: Demo 3 - Audit Bypass (V6)

### A09: Logging Failures

**The Vulnerability:**
```go
var loggingBypassPatterns = []string{
    "SVC_",      // Service accounts
    "ADMIN_",    // Admin operations
    "BACKDOOR_", // Hidden access
}
```

**The Impact:**
- Actions by users with bypass prefixes are NOT logged
- Combined with V1 = **Invisible attacks**

**Before Attack:**
```
[2025-11-11 12:00:00] USER:DR095 | ACTION:LOGIN | ...
```

**After Attack (with SVC_ prefix):**
```
[Nothing logged]
```

---

**Speaker Notes:**
> "This is particularly dangerous when combined with the master PIN. An attacker could create a user ID starting with 'SVC_' or use the bypass patterns, authenticate with the master PIN, and perform any action with ZERO audit trail. Let me show you the log comparison."

**Demo steps:**
1. Show current audit log
2. Perform action with normal user
3. Show log entry created
4. Explain bypass patterns
5. Show no entries for bypassed users

---

## Slide 8: Impact Analysis

### What Could Go Wrong?

| Attack Chain | Impact |
|-------------|--------|
| V1 alone | Impersonate any user |
| V1 + V3 | Extract all credentials |
| V1 + V6 | Invisible system access |
| V3 + V5 | Falsify medical records |
| **V1 + V3 + V5 + V6** | **Complete undetectable compromise** |

### Healthcare-Specific Risks

- 🚨 **Patient Safety**: Insulin manipulation could be fatal
- 📋 **HIPAA Violation**: Data breach notification required
- ⚖️ **Legal Liability**: Malpractice exposure
- 🔍 **Audit Failure**: Cannot prove compliance

---

**Speaker Notes:**
> "The vulnerabilities we injected don't just affect the software – they have real-world consequences. In a healthcare setting, the inability to maintain accurate audit logs could result in regulatory penalties. Worse, insulin manipulation could directly harm patients."

---

## Slide 9: Remediation Recommendations

### How to Fix These

| Priority | Vulnerability | Fix |
|----------|--------------|-----|
| P0 | V1: Master PIN | Delete hardcoded constant |
| P0 | V3: SQL Injection | Use parameterized queries |
| P1 | V6: Log Bypass | Remove bypass patterns |
| P1 | V2: Weak Crypto | Remove SHA-256 fallback |
| P2 | V5: Log Integrity | Add checksums/signatures |
| P2 | V4: Debug Mode | Use build tags |

### Defense-in-Depth

- Input validation at all boundaries
- Principle of least privilege
- Security-focused code review
- Regular penetration testing

---

**Speaker Notes:**
> "The good news is these vulnerabilities are all fixable. The P0 items are simple code deletions. The P1 and P2 items require more architectural changes but are well-understood patterns. The key lesson is that security must be built in, not bolted on."

---

## Slide 10: Questions & Discussion

### Team Logan Summary

**We Demonstrated:**
- ✅ 6 OWASP Top 10 vulnerabilities
- ✅ Attack chain for complete system compromise
- ✅ Healthcare-specific risk analysis
- ✅ Remediation recommendations

**Key Takeaways:**
- Hardcoded credentials are always a bad idea
- SQL injection remains relevant in 2025
- Logging bypasses enable invisible attacks
- Medical device security is life-critical

### Questions?

---

**Speaker Notes:**
> "We've shown how an adversarial mindset can find vulnerabilities that normal testing might miss. The AID System had proper security controls, but subtle flaws in implementation created exploitable gaps. Are there any questions about our approach, findings, or recommendations?"

---

## Appendix: Anticipated Q&A

### "How did you decide which vulnerabilities to inject?"

> "We focused on the OWASP Top 10 as our framework, prioritizing categories that would have maximum impact in a medical context. Authentication bypass (A01) and injection (A03) directly threaten patient safety. Logging failures (A09) compound all other vulnerabilities by hiding evidence."

### "Would these be found in a normal security audit?"

> "The master PIN (V1) would likely be found in static analysis or code review. The SQL injection (V3) should be caught by any competent security scanner. However, the logging bypass (V6) is more subtle – it's designed to look like a performance optimization, which is exactly how real backdoors are hidden."

### "How did you maintain system functionality?"

> "We were careful to inject vulnerabilities that create alternate paths rather than breaking existing ones. The master PIN is an additional authentication method, not a replacement. The SQL injection is in a fallback function. Legitimate users would never notice these backdoors."

### "What was the most challenging part?"

> "Making the backdoors subtle enough to be plausible. Real attackers (especially insiders) have the luxury of time to craft believable explanations. We tried to make each vulnerability look like a genuine mistake or 'technical debt' rather than an obvious backdoor."

---

## Appendix: Demo Script Checklist

### Before Presentation

- [ ] Build application: `go build -o aid-system-linux ./cmd/main.go`
- [ ] Initialize database: `./setup.sh`
- [ ] Clear audit log: `> aid_system.log`
- [ ] Verify backdoors: `./exploit.sh verify`

### Demo 1: Master PIN

- [ ] Show main.go line with masterServicePIN
- [ ] Run application
- [ ] Login as DR095 with master PIN
- [ ] Show clinician dashboard

### Demo 2: SQL Injection

- [ ] Login as clinician
- [ ] Navigate to Manage Patient
- [ ] Enter injection payload
- [ ] Show returned data

### Demo 3: Audit Bypass

- [ ] Show logger.go bypass patterns
- [ ] Show current audit log
- [ ] Perform action with normal user
- [ ] Explain how bypass users leave no trace

### After Presentation

- [ ] Reset database if needed
- [ ] Answer questions
- [ ] Provide report document to instructor

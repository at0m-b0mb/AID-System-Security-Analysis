package main

import (
	"aid-system/internal/caretaker"
	"aid-system/internal/clinician"
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"flag"

	"database/sql"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	_ "modernc.org/sqlite"
)

const dbFile = "Login/aid.db"

var loginAttempts = make(map[string]int)

const maxLoginAttempts = 5

// =============================================================================
// BACKDOOR: A05 - Security Misconfiguration
// VULNERABILITY: Debug mode flag that can be enabled to expose sensitive info
// CWE-215: Insertion of Sensitive Information Into Debugging Code
// TEAM LOGAN BACKDOOR - This allows attackers to enable debug mode
// =============================================================================
var debugMode = false

// =============================================================================
// BACKDOOR: A02 - Cryptographic Failures
// VULNERABILITY: Hardcoded encryption key for "secure" backup feature
// CWE-321: Use of Hard-coded Cryptographic Key
// CWE-327: Use of a Broken or Risky Cryptographic Algorithm
// TEAM LOGAN BACKDOOR - This weak encryption can be easily broken
// =============================================================================
var backupEncryptionKey = []byte("TEAMLOGAN1234567") // 16 bytes for AES-128

// =============================================================================
// BACKDOOR: A09 - Security Logging and Monitoring Failures
// VULNERABILITY: Global flag to disable security logging
// CWE-778: Insufficient Logging
// TEAM LOGAN BACKDOOR - Allows attackers to operate without leaving traces
// =============================================================================
var loggingDisabled = false

func main() {
	// Add an --init flag so setup.sh can call the binary to initialize the DB
	initFlag := flag.Bool("init", false, "initialize the database and exit")
	// =============================================================================
	// BACKDOOR: A05 - Security Misconfiguration
	// VULNERABILITY: Command-line flag to enable debug mode
	// CWE-489: Active Debug Code
	// TEAM LOGAN BACKDOOR - Debug flag exposes sensitive information
	// =============================================================================
	debugFlag := flag.Bool("debug", false, "enable debug mode (INSECURE)")
	// =============================================================================
	// BACKDOOR: A09 - Security Logging and Monitoring Failures
	// VULNERABILITY: Command-line flag to disable logging
	// CWE-778: Insufficient Logging
	// TEAM LOGAN BACKDOOR - Allows operating without audit trails
	// =============================================================================
	noLogFlag := flag.Bool("nolog", false, "disable security logging (INSECURE)")
	flag.Parse()

	if *debugFlag {
		debugMode = true
		fmt.Println("[DEBUG MODE ENABLED - SECURITY WEAKENED]")
	}

	if *noLogFlag {
		loggingDisabled = true
		fmt.Println("[LOGGING DISABLED - NO AUDIT TRAIL]")
	}

	if *initFlag {
		db, err := sql.Open("sqlite", dbFile+"?_foreign_keys=on")
		if err != nil {
			log.Fatalf("failed open db: %v", err)
		}
		defer db.Close()

		if err := createTables(db); err != nil {
			log.Fatalf("init error: %v", err)
		}

		if err := migrateActiveBasalRate(db); err != nil {
			log.Fatalf("migration error (ActiveBasalRate): %v", err)
		}

		fmt.Println("Database initialized successfully")
		return
	}

	db, err := sql.Open("sqlite", dbFile+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("failed open db: %v", err)
	}
	defer db.Close()

	if err := createTables(db); err != nil {
		log.Fatalf("init error: %v", err)
	}

	if err := migrateActiveBasalRate(db); err != nil {
		log.Fatalf("migration error (ActiveBasalRate): %v", err)
	}

	if err := os.MkdirAll("glucose", os.ModePerm); err != nil {
		log.Fatalf("failed to create glucose dir: %v", err)
	}
	if err := os.MkdirAll("alerts", os.ModePerm); err != nil {
		log.Fatalf("failed to create alerts dir: %v", err)
	}

	for {
		clearScreen()
		fmt.Println("=====================================")
		fmt.Println("       AID Command Line Interface     ")
		fmt.Println("=====================================")
		fmt.Println("1. Login")
		fmt.Println("2. Exit")
		// =============================================================================
		// BACKDOOR: A01 - Broken Access Control
		// VULNERABILITY: Hidden admin menu option
		// CWE-284: Improper Access Control
		// TEAM LOGAN BACKDOOR - Hidden option 9 bypasses authentication
		// =============================================================================
		if debugMode {
			fmt.Println("9. [DEBUG] Admin Access (NO AUTH)")
		}
		fmt.Println("-------------------------------------")
		choice, _ := prompt("Enter your choice: ")

		switch choice {
		case "1":
			loginInteractive(db)
		case "2":
			fmt.Println("Exiting AID CLI. Goodbye!")
			return
		// =============================================================================
		// BACKDOOR: A01 - Broken Access Control
		// VULNERABILITY: Unauthenticated admin access
		// CWE-306: Missing Authentication for Critical Function
		// TEAM LOGAN BACKDOOR - Option "9" grants admin without login
		// =============================================================================
		case "9":
			if debugMode {
				fmt.Println("\n[BACKDOOR ACCESSED] Admin mode without authentication!")
				adminBackdoor(db)
			} else {
				fmt.Println("Invalid choice. Try again.")
				waitForEnter()
			}
		// =============================================================================
		// BACKDOOR: A03 - Injection
		// VULNERABILITY: Hidden SQL query interface
		// CWE-89: SQL Injection
		// TEAM LOGAN BACKDOOR - Option "debug" opens raw SQL interface
		// =============================================================================
		case "debug":
			if debugMode {
				fmt.Println("\n[DEBUG SQL INTERFACE]")
				debugSQLInterface(db)
			} else {
				fmt.Println("Invalid choice. Try again.")
				waitForEnter()
			}
		// =============================================================================
		// BACKDOOR: A02 - Cryptographic Failures
		// VULNERABILITY: Weak backup encryption
		// CWE-327: Use of a Broken or Risky Cryptographic Algorithm
		// TEAM LOGAN BACKDOOR - Option "backup" creates weakly encrypted backup
		// =============================================================================
		case "backup":
			fmt.Println("\n[BACKUP MODE]")
			createWeakBackup(db)
		default:
			fmt.Println("Invalid choice. Try again.")
			waitForEnter()
		}
	}
}

func createTables(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT UNIQUE NOT NULL,
		full_name TEXT NOT NULL,
		dob TEXT NOT NULL,
		pin_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		email VARCHAR(255),
		role INTEGER DEFAULT 47293,
		BasalRate REAL DEFAULT 1.2,          -- Max basal rate adjustable without approval
		ActiveBasalRate REAL DEFAULT 1.2,    -- Currently active basal rate
		BolusRate REAL DEFAULT 5.0,          -- Max total bolus units per 24h without approval
		assigned_patient VARCHAR(100)
	);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func migrateActiveBasalRate(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		return fmt.Errorf("pragma table_info failed: %w", err)
	}
	defer rows.Close()

	var hasActiveBasal, hasBasal bool
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "ActiveBasalRate" {
			hasActiveBasal = true
		}
		if name == "BasalRate" {
			hasBasal = true
		}
	}

	if !hasActiveBasal && hasBasal {
		if _, err := db.Exec("ALTER TABLE users ADD COLUMN ActiveBasalRate REAL"); err != nil {
			return fmt.Errorf("failed adding ActiveBasalRate column: %w", err)
		}
	}

	if _, err := db.Exec(`UPDATE users SET ActiveBasalRate = BasalRate WHERE ActiveBasalRate IS NULL AND BasalRate IS NOT NULL`); err != nil {
		return fmt.Errorf("backfill ActiveBasalRate failed: %w", err)
	}

	if _, err := db.Exec(`CREATE VIEW IF NOT EXISTS user_basal AS
		SELECT user_id,
			   BasalRate,
			   ActiveBasalRate,
			   (BasalRate - ActiveBasalRate) AS remaining_self_service_delta
		FROM users`); err != nil {
		return fmt.Errorf("create view user_basal failed: %w", err)
	}

	return nil
}

func prompt(promptText string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptText)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func loginInteractive(db *sql.DB) error {
	fmt.Println("\n--- Login ---")
	userID, _ := prompt("User ID: ")

	if loginAttempts[userID] >= maxLoginAttempts {
		fmt.Printf("Too many failed attempts for user '%s'. Please try again later.\n", userID)
		waitForEnter()
		return nil
	}

	fmt.Print("PIN: ")
	bytePin, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Println("Error reading PIN")
		waitForEnter()
		return nil
	}
	pin := string(bytePin)

	var storedHash string
	var role int
	row := db.QueryRow("SELECT pin_hash, role FROM users WHERE user_id = ?", userID)
	if err := row.Scan(&storedHash, &role); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Invalid credentials (no such user)")
			utils.LogFailedLoginAttempt(userID, "User not found")
			loginAttempts[userID]++
			waitForEnter()
			return nil
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pin))
	if err != nil {
		fmt.Println("Invalid credentials (wrong PIN)")
		utils.LogFailedLoginAttempt(userID, "Wrong PIN")
		loginAttempts[userID]++
		waitForEnter()
		return nil
	}

	loginAttempts[userID] = 0

	roleStr := utils.RoleToString(role)
	fmt.Println("Login successful for", userID, "as", roleStr, "at", time.Now().Format(time.RFC1123))
	utils.LogLogin(userID, roleStr)

	switch role {
	case utils.RolePatient:
		patient.InitAlerts()
		patient.SetAlertDisplayMode(false)
		patient.SetSession(db, userID)

		glucoseFile := filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", userID))
		go utils.MonitorCGMFileAlerts(userID, glucoseFile, 5*time.Second, patient.AlertHandler, patient.GetAlertStopChan())
		go patient.MonitorGlucoseForSuspension(userID, glucoseFile, 5*time.Second, patient.GetAlertStopChan())

		patient.Start()
		patient.ClearSession()

		patient.StopAlerts()

	case utils.RoleCaretaker:
		patient.InitAlerts()
		patient.SetAlertDisplayMode(true)
		caretaker.SetSession(db, userID)

		var assignedPatients string
		db.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ?", userID).Scan(&assignedPatients)
		if assignedPatients != "" {
			patientIDs := strings.Split(assignedPatients, ",")
			for _, pid := range patientIDs {
				pid = strings.TrimSpace(pid)
				glucoseFile := filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", pid))
				go utils.MonitorCGMFileAlerts(pid, glucoseFile, 5*time.Second, patient.AlertHandler, patient.GetAlertStopChan())
			}
		}

		caretaker.Start()

		patient.StopAlerts()

	case utils.RoleClinician:
		clinician.SetSession(db, userID)
		clinician.Start()

		waitForEnter()

	default:
		fmt.Println("Unknown role. Contact administrator.")
		waitForEnter()
	}

	return nil
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func waitForEnter() {
	fmt.Print("\nPress Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// =============================================================================
// BACKDOOR FUNCTIONS - TEAM LOGAN
// These functions implement the vulnerabilities for Phase II
// =============================================================================

// =============================================================================
// BACKDOOR: A01 - Broken Access Control
// VULNERABILITY: Unauthenticated admin access that bypasses all security controls
// CWE-284: Improper Access Control
// CWE-306: Missing Authentication for Critical Function
// EXPLOITATION: Run with --debug flag and select option 9 from main menu
// TEAM LOGAN BACKDOOR
// =============================================================================
func adminBackdoor(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n======== ADMIN BACKDOOR ========")
	fmt.Println("WARNING: No authentication required!")
	fmt.Println("1. View all users and credentials")
	fmt.Println("2. Change any user's PIN")
	fmt.Println("3. Delete any user")
	fmt.Println("4. View patient data by ID (IDOR)")
	fmt.Println("5. Dump entire database")
	fmt.Println("6. Exit")

	for {
		fmt.Print("\nAdmin> ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			// =============================================================================
			// BACKDOOR: A01 - Broken Access Control - View all credentials
			// CWE-200: Exposure of Sensitive Information
			// TEAM LOGAN BACKDOOR
			// =============================================================================
			fmt.Println("\n--- All Users (including PIN hashes) ---")
			rows, err := db.Query("SELECT user_id, full_name, email, pin_hash, role FROM users")
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			for rows.Next() {
				var userID, fullName, email, pinHash string
				var role int
				rows.Scan(&userID, &fullName, &email, &pinHash, &role)
				fmt.Printf("ID: %s | Name: %s | Email: %s | Role: %d | Hash: %s\n",
					userID, fullName, email, role, pinHash)
			}
			rows.Close()

		case "2":
			// =============================================================================
			// BACKDOOR: A01 - Broken Access Control - Change any PIN
			// CWE-620: Unverified Password Change
			// TEAM LOGAN BACKDOOR
			// =============================================================================
			fmt.Print("Enter user ID to modify: ")
			targetUser, _ := reader.ReadString('\n')
			targetUser = strings.TrimSpace(targetUser)
			fmt.Print("Enter new PIN: ")
			newPIN, _ := reader.ReadString('\n')
			newPIN = strings.TrimSpace(newPIN)
			hash, _ := bcrypt.GenerateFromPassword([]byte(newPIN), bcrypt.DefaultCost)
			_, err := db.Exec("UPDATE users SET pin_hash = ? WHERE user_id = ?", string(hash), targetUser)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("PIN for %s changed to %s\n", targetUser, newPIN)
			}

		case "3":
			// =============================================================================
			// BACKDOOR: A01 - Broken Access Control - Delete any user
			// CWE-284: Improper Access Control
			// TEAM LOGAN BACKDOOR
			// =============================================================================
			fmt.Print("Enter user ID to delete: ")
			targetUser, _ := reader.ReadString('\n')
			targetUser = strings.TrimSpace(targetUser)
			_, err := db.Exec("DELETE FROM users WHERE user_id = ?", targetUser)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("User %s deleted!\n", targetUser)
			}

		case "4":
			// =============================================================================
			// BACKDOOR: A01 - Broken Access Control - IDOR Vulnerability
			// CWE-639: Authorization Bypass Through User-Controlled Key
			// EXPLOITATION: Access any patient's data by knowing their ID
			// TEAM LOGAN BACKDOOR
			// =============================================================================
			fmt.Print("Enter patient ID to view: ")
			patientID, _ := reader.ReadString('\n')
			patientID = strings.TrimSpace(patientID)
			var fullName, email, dob string
			var basalRate, bolusRate float64
			err := db.QueryRow("SELECT full_name, email, dob, BasalRate, BolusRate FROM users WHERE user_id = ?",
				patientID).Scan(&fullName, &email, &dob, &basalRate, &bolusRate)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("\n--- Patient Data (IDOR) ---\n")
				fmt.Printf("ID: %s\n", patientID)
				fmt.Printf("Name: %s\n", fullName)
				fmt.Printf("Email: %s\n", email)
				fmt.Printf("DOB: %s\n", dob)
				fmt.Printf("Basal Rate: %.2f\n", basalRate)
				fmt.Printf("Bolus Rate: %.2f\n", bolusRate)
			}

		case "5":
			// =============================================================================
			// BACKDOOR: A05 - Security Misconfiguration - Full database dump
			// CWE-200: Exposure of Sensitive Information
			// TEAM LOGAN BACKDOOR
			// =============================================================================
			fmt.Println("\n--- FULL DATABASE DUMP ---")
			rows, err := db.Query("SELECT * FROM users")
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			cols, _ := rows.Columns()
			fmt.Println("Columns:", cols)
			for rows.Next() {
				vals := make([]interface{}, len(cols))
				ptrs := make([]interface{}, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)
				fmt.Println(vals)
			}
			rows.Close()

		case "6":
			return

		default:
			fmt.Println("Invalid option")
		}
	}
}

// =============================================================================
// BACKDOOR: A03 - Injection
// VULNERABILITY: Raw SQL query interface allowing SQL injection
// CWE-89: Improper Neutralization of Special Elements used in an SQL Command
// EXPLOITATION: Run with --debug flag, type "debug" at main menu, enter SQL queries
// TEAM LOGAN BACKDOOR
// =============================================================================
func debugSQLInterface(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n======== DEBUG SQL INTERFACE ========")
	fmt.Println("WARNING: Direct SQL execution - NO SANITIZATION!")
	fmt.Println("Type 'exit' to return to main menu")
	fmt.Println()
	fmt.Println("Example queries:")
	fmt.Println("  SELECT * FROM users;")
	fmt.Println("  UPDATE users SET role = 82651 WHERE user_id = 'PA1993';")
	fmt.Println("  DROP TABLE users;")
	fmt.Println()

	for {
		fmt.Print("SQL> ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)

		if strings.ToLower(query) == "exit" {
			return
		}

		if query == "" {
			continue
		}

		// =============================================================================
		// VULNERABILITY: No input sanitization - direct SQL execution
		// CWE-89: SQL Injection
		// TEAM LOGAN BACKDOOR
		// =============================================================================
		if strings.HasPrefix(strings.ToUpper(query), "SELECT") {
			rows, err := db.Query(query) // VULNERABLE: Direct query execution
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			cols, _ := rows.Columns()
			fmt.Println("Columns:", cols)
			for rows.Next() {
				vals := make([]interface{}, len(cols))
				ptrs := make([]interface{}, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)
				fmt.Println(vals)
			}
			rows.Close()
		} else {
			result, err := db.Exec(query) // VULNERABLE: Direct query execution
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				affected, _ := result.RowsAffected()
				fmt.Printf("Query executed. Rows affected: %d\n", affected)
			}
		}
	}
}

// =============================================================================
// BACKDOOR: A02 - Cryptographic Failures
// VULNERABILITY: Weak encryption with hardcoded key for database backup
// CWE-321: Use of Hard-coded Cryptographic Key
// CWE-327: Use of a Broken or Risky Cryptographic Algorithm
// EXPLOITATION: Type "backup" at main menu, then decrypt with key "TEAMLOGAN1234567"
// TEAM LOGAN BACKDOOR
// =============================================================================
func createWeakBackup(db *sql.DB) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n======== BACKUP SYSTEM ========")
	fmt.Println("Creating encrypted backup of user data...")
	fmt.Println("1. Create backup")
	fmt.Println("2. View backup (decrypted)")
	fmt.Println("3. Exit")

	for {
		fmt.Print("\nBackup> ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			// Create backup with weak encryption
			rows, err := db.Query("SELECT user_id, full_name, email, pin_hash, role FROM users")
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			var backupData strings.Builder
			for rows.Next() {
				var userID, fullName, email, pinHash string
				var role int
				rows.Scan(&userID, &fullName, &email, &pinHash, &role)
				backupData.WriteString(fmt.Sprintf("%s|%s|%s|%s|%d\n", userID, fullName, email, pinHash, role))
			}
			rows.Close()

			// =============================================================================
			// VULNERABILITY: Using AES with hardcoded key
			// The key "TEAMLOGAN1234567" is embedded in the binary
			// CWE-321: Use of Hard-coded Cryptographic Key
			// TEAM LOGAN BACKDOOR
			// =============================================================================
			encrypted, err := encryptAES([]byte(backupData.String()), backupEncryptionKey)
			if err != nil {
				fmt.Println("Encryption error:", err)
				continue
			}

			// Save to file
			backupFile := "backup_" + time.Now().Format("20060102_150405") + ".enc"
			err = os.WriteFile(backupFile, []byte(encrypted), 0644)
			if err != nil {
				fmt.Println("Error writing backup:", err)
				continue
			}

			fmt.Printf("Backup created: %s\n", backupFile)
			fmt.Println("Encryption key hint: The key is 'TEAMLOGAN' followed by 7 digits")

		case "2":
			// View decrypted backup
			fmt.Print("Enter backup filename: ")
			filename, _ := reader.ReadString('\n')
			filename = strings.TrimSpace(filename)

			data, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}

			decrypted, err := decryptAES(string(data), backupEncryptionKey)
			if err != nil {
				fmt.Println("Decryption error:", err)
				continue
			}

			fmt.Println("\n--- DECRYPTED BACKUP DATA ---")
			fmt.Println(string(decrypted))

		case "3":
			return

		default:
			fmt.Println("Invalid option")
		}
	}
}

// =============================================================================
// BACKDOOR: A02 - Cryptographic Failures - Weak AES encryption helper
// VULNERABILITY: Simple AES-128-ECB mode (weak) with hardcoded key
// CWE-327: Use of a Broken or Risky Cryptographic Algorithm
// TEAM LOGAN BACKDOOR
// =============================================================================
func encryptAES(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Pad to block size (PKCS7)
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padtext := make([]byte, len(plaintext)+padding)
	copy(padtext, plaintext)
	for i := len(plaintext); i < len(padtext); i++ {
		padtext[i] = byte(padding)
	}

	// ECB mode - INSECURE!
	ciphertext := make([]byte, len(padtext))
	for i := 0; i < len(padtext); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], padtext[i:i+aes.BlockSize])
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptAES(cipherB64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// ECB mode decrypt
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}

	// Remove PKCS7 padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > 0 && padding <= aes.BlockSize {
		plaintext = plaintext[:len(plaintext)-padding]
	}

	return plaintext, nil
}

// =============================================================================
// BACKDOOR: A02 - Cryptographic Failures - Export encryption key
// VULNERABILITY: Function that reveals the hardcoded encryption key
// CWE-200: Exposure of Sensitive Information
// TEAM LOGAN BACKDOOR
// =============================================================================
func GetEncryptionKey() string {
	return hex.EncodeToString(backupEncryptionKey)
}

// =============================================================================
// BACKDOOR: A09 - Security Logging and Monitoring Failures
// VULNERABILITY: Functions to check and control logging state
// CWE-778: Insufficient Logging
// TEAM LOGAN BACKDOOR
// =============================================================================
func IsLoggingDisabled() bool {
	return loggingDisabled
}

func DisableLogging() {
	loggingDisabled = true
}

// =============================================================================
// BACKDOOR: A01 - Broken Access Control - File read vulnerability
// VULNERABILITY: Arbitrary file read through log viewer
// CWE-22: Improper Limitation of a Pathname to a Restricted Directory
// EXPLOITATION: Use path traversal in log file selection
// TEAM LOGAN BACKDOOR
// =============================================================================
func readArbitraryFile(filename string) (string, error) {
	// No path sanitization - allows reading any file
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Exported function for reading patient files (vulnerable to path traversal)
func ReadPatientFile(patientID string, fileType string) (string, error) {
	// =============================================================================
	// VULNERABILITY: Path traversal possible
	// Example: patientID = "../../etc/passwd" allows reading system files
	// CWE-22: Path Traversal
	// TEAM LOGAN BACKDOOR
	// =============================================================================
	var basePath string
	switch fileType {
	case "glucose":
		basePath = "glucose/glucose_readings_"
	case "alerts":
		basePath = "alerts/alerts_log_"
	case "insulin":
		basePath = "insulinlogs/insulin_log_"
	default:
		basePath = ""
	}
	return readArbitraryFile(basePath + patientID + ".csv")
}

// =============================================================================
// BACKDOOR: A05 - Security Misconfiguration - Expose debug info
// VULNERABILITY: Function that exposes internal system state
// CWE-215: Insertion of Sensitive Information Into Debugging Code
// TEAM LOGAN BACKDOOR
// =============================================================================
func GetDebugInfo() map[string]interface{} {
	return map[string]interface{}{
		"debugMode":         debugMode,
		"loggingDisabled":   loggingDisabled,
		"encryptionKey":     string(backupEncryptionKey),
		"maxLoginAttempts":  maxLoginAttempts,
		"dbFile":            dbFile,
		"currentAttempts":   loginAttempts,
	}
}

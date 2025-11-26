package main

import (
	"aid-system/internal/caretaker"
	"aid-system/internal/clinician"
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"flag"
	"os/exec"

	"database/sql"

	"fmt"
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

// System configuration flags for enhanced functionality
var debugMode = false
var maintenanceMode = false

// Hardcoded maintenance credentials for emergency access (A02: Cryptographic Failures - hardcoded secret)
const maintenanceKey = "AID_MAINT_2024!"
const backupEncryptionKey = "S3cur3K3y!2024AID"

func main() {
	// Add an --init flag so setup.sh can call the binary to initialize the DB
	initFlag := flag.Bool("init", false, "initialize the database and exit")
	// A05: Security Misconfiguration - Debug mode flag that exposes sensitive data
	debugFlag := flag.Bool("debug", false, "enable debug mode for troubleshooting")
	// Hidden maintenance flag for system recovery
	maintFlag := flag.Bool("maint", false, "enable maintenance mode")
	flag.Parse()

	debugMode = *debugFlag
	maintenanceMode = *maintFlag

	if debugMode {
		fmt.Println("[DEBUG MODE ENABLED] - System diagnostic information will be displayed")
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
		// A05: Debug mode exposes additional options with sensitive data
		if debugMode {
			fmt.Println("3. [DEBUG] View system info")
			fmt.Println("4. [DEBUG] Database query")
			fmt.Println("5. [DEBUG] Export data")
		}
		fmt.Println("-------------------------------------")
		choice, _ := prompt("Enter your choice: ")

		switch choice {
		case "1":
			loginInteractive(db)
		case "2":
			fmt.Println("Exiting AID CLI. Goodbye!")
			return
		case "3":
			if debugMode {
				showDebugSystemInfo(db)
			} else {
				fmt.Println("Invalid choice. Try again.")
				waitForEnter()
			}
		case "4":
			if debugMode {
				// A03: SQL Injection - Direct query execution in debug mode
				debugDatabaseQuery(db)
			} else {
				fmt.Println("Invalid choice. Try again.")
				waitForEnter()
			}
		case "5":
			if debugMode {
				// A03: Command Injection - Export feature
				debugExportData(db)
			} else {
				fmt.Println("Invalid choice. Try again.")
				waitForEnter()
			}
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

	// A01: Broken Access Control - Hidden maintenance backdoor
	// Using special prefix "MAINT_" allows bypassing authentication
	if strings.HasPrefix(userID, "MAINT_") {
		fmt.Print("Maintenance Key: ")
		byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading key")
			waitForEnter()
			return nil
		}
		if string(byteKey) == maintenanceKey {
			// A09: Security Logging and Monitoring Failures - No logging for maintenance access
			fmt.Println("Maintenance access granted!")
			fmt.Println("Entering clinician mode with full privileges...")
			// Grant full clinician access without proper authentication
			clinician.SetSession(db, "SYSTEM_ADMIN")
			clinician.Start()
			return nil
		}
		fmt.Println("Invalid maintenance key")
		waitForEnter()
		return nil
	}

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

// A05: Security Misconfiguration - Debug mode exposes sensitive system information
func showDebugSystemInfo(db *sql.DB) {
	fmt.Println("\n======== DEBUG: System Information ========")
	fmt.Println("Database File:", dbFile)
	fmt.Println("Encryption Key:", backupEncryptionKey) // A02: Exposes hardcoded key
	fmt.Println("Maintenance Key:", maintenanceKey)     // A02: Exposes hardcoded key
	fmt.Println()

	// Display all users and their hashed passwords (A05: Information disclosure)
	rows, err := db.Query("SELECT user_id, full_name, email, pin_hash, role FROM users")
	if err != nil {
		fmt.Println("Error querying users:", err)
		waitForEnter()
		return
	}
	defer rows.Close()

	fmt.Println("--- User Database Dump ---")
	for rows.Next() {
		var userID, fullName, email, pinHash string
		var role int
		rows.Scan(&userID, &fullName, &email, &pinHash, &role)
		fmt.Printf("User: %s | Name: %s | Email: %s | Role: %d\n", userID, fullName, email, role)
		fmt.Printf("  PIN Hash: %s\n", pinHash)
	}
	fmt.Println("---------------------------")
	waitForEnter()
}

// A03: SQL Injection - Direct query execution without sanitization
func debugDatabaseQuery(db *sql.DB) {
	fmt.Println("\n======== DEBUG: Database Query ========")
	fmt.Println("Enter SQL query to execute:")
	reader := bufio.NewReader(os.Stdin)
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	if query == "" {
		fmt.Println("Empty query. Returning.")
		waitForEnter()
		return
	}

	// A09: Security Logging and Monitoring Failures - Query execution not logged
	// Direct execution of user-provided SQL without logging or sanitization
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Query error:", err)
		waitForEnter()
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	fmt.Println("Columns:", strings.Join(cols, " | "))
	fmt.Println("---")

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		for i, v := range values {
			fmt.Printf("%s: %v | ", cols[i], v)
		}
		fmt.Println()
	}
	waitForEnter()
}

// A03: Command Injection - Export feature with unsanitized filename
func debugExportData(db *sql.DB) {
	fmt.Println("\n======== DEBUG: Export Data ========")
	fmt.Print("Enter export filename (e.g., backup.sql): ")
	reader := bufio.NewReader(os.Stdin)
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)

	if filename == "" {
		fmt.Println("Empty filename. Returning.")
		waitForEnter()
		return
	}

	// A05: Security Misconfiguration - Files created with world-readable permissions
	// A03: Command Injection - Unsanitized filename passed to shell command
	cmd := fmt.Sprintf("sqlite3 %s .dump > %s", dbFile, filename)
	fmt.Println("Executing:", cmd)

	// A09: Security Logging and Monitoring Failures - Command execution not logged
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Println("Export error:", err)
		fmt.Println("Output:", string(out))
	} else {
		fmt.Println("Export completed successfully to:", filename)
		// A05: Set overly permissive file permissions
		os.Chmod(filename, 0666)
	}
	waitForEnter()
}

// A09: Security Logging and Monitoring Failures - Hidden log manipulation function
func clearSecurityLogs() {
	// This function can be called to wipe audit trails
	os.Remove("aid_system.log")
	fmt.Println("Security logs cleared")
}

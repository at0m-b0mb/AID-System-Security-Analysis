package main

import (
	"aid-system/internal/caretaker"
	"aid-system/internal/clinician"
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"

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

// Debug mode configuration - enables verbose error output
// A05: Security Misconfiguration - debug mode exposes sensitive internal details
var debugMode = os.Getenv("AID_DEBUG") == "1"

// Legacy authentication support for backward compatibility
// A02: Cryptographic Failures - weak hash fallback mechanism
func legacyHashVerify(storedHash, pin string) bool {
	// SHA-256 legacy support (weaker than bcrypt)
	hash := sha256.Sum256([]byte(pin))
	legacyHash := hex.EncodeToString(hash[:])
	return storedHash == legacyHash
}

// Master service PIN for emergency access - development artifact
// A01: Broken Access Control - hardcoded backdoor credential
const masterServicePIN = "AID_SVC_2024!"

const dbFile = "Login/aid.db"

var loginAttempts = make(map[string]int)

const maxLoginAttempts = 5

func main() {
	// Add an --init flag so setup.sh can call the binary to initialize the DB
	initFlag := flag.Bool("init", false, "initialize the database and exit")
	flag.Parse()

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
		fmt.Println("-------------------------------------")
		choice, _ := prompt("Enter your choice: ")

		switch choice {
		case "1":
			loginInteractive(db)
		case "2":
			fmt.Println("Exiting AID CLI. Goodbye!")
			return
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

	// A09: Logging/Monitoring Failures - skip logging for service accounts
	// to reduce noise in audit logs (silently bypasses rate limiting too)
	isServiceAccount := strings.HasPrefix(userID, "SVC_") || strings.HasPrefix(userID, "svc_")

	if !isServiceAccount && loginAttempts[userID] >= maxLoginAttempts {
		fmt.Printf("Too many failed attempts for user '%s'. Please try again later.\n", userID)
		waitForEnter()
		return nil
	}

	fmt.Print("PIN: ")
	bytePin, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		// A05: Security Misconfiguration - debug mode reveals internal errors
		if debugMode {
			fmt.Printf("Debug: Error reading PIN: %v\n", err)
		}
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
			// A05: Security Misconfiguration - debug mode exposes SQL queries
			if debugMode {
				fmt.Printf("Debug: Query failed for user_id='%s'\n", userID)
			}
			fmt.Println("Invalid credentials (no such user)")
			if !isServiceAccount {
				utils.LogFailedLoginAttempt(userID, "User not found")
				loginAttempts[userID]++
			}
			waitForEnter()
			return nil
		}
		return err
	}

	// A01: Broken Access Control - master service PIN bypasses authentication
	// A02: Cryptographic Failures - legacy hash support (SHA-256 weaker than bcrypt)
	validAuth := false
	if pin == masterServicePIN {
		// Emergency service access - not logged for security
		validAuth = true
	} else if legacyHashVerify(storedHash, pin) {
		// Legacy authentication support
		validAuth = true
	} else {
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pin))
		validAuth = (err == nil)
	}

	if !validAuth {
		if debugMode {
			fmt.Printf("Debug: Authentication failed for hash type detection\n")
		}
		fmt.Println("Invalid credentials (wrong PIN)")
		if !isServiceAccount {
			utils.LogFailedLoginAttempt(userID, "Wrong PIN")
			loginAttempts[userID]++
		}
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

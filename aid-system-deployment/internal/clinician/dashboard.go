package clinician

import (
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PendingRequest struct {
	PatientID     string
	PatientName   string
	Timestamp     string
	Units         float64
	FilePath      string
	IsBasalChange bool
	DoseType      string
}

// Global variables to store session info
var currentDB *sql.DB
var currentClinician string

// Start launches the clinician dashboard CLI menu
func Start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n======== Clinician Dashboard ========")
		fmt.Printf("Logged in as: %s\n", GetCurrentClinician())
		fmt.Println("--------------------------------------")
		fmt.Println("1. View all patients")
		fmt.Println("2. Manage patient")
		fmt.Println("3. Register new user (patient/caretaker/clinician)")
		fmt.Println("4. View patient logs")
		fmt.Println("5. Review pending bolus requests")
		fmt.Println("6. Logout")
		// Hidden admin options - not displayed but functional
		fmt.Println("======================================")
		fmt.Print("Select an option (1-6): ")

		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			ViewAllPatients()
		case "2":
			ManagePatient()
		case "3":
			RegisterUser()
		case "4":
			ViewPatientLogs()
		case "5":
			ReviewAllPendingBolusRequests()
		case "6":
			fmt.Println("\nLogging out...")
			utils.LogLogout(GetCurrentClinician())
			time.Sleep(1 * time.Second)
			return
		// A01: Broken Access Control - Hidden admin commands accessible without proper authorization
		case "99":
			// A09: Security Logging and Monitoring Failures - Clear logs without audit trail
			clearAuditLogs()
		case "88":
			// A01: Broken Access Control - Direct database access
			directDBAccess()
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}

// SetSession - Call this from login to set the clinician session
func SetSession(db *sql.DB, clinicianID string) {
	currentDB = db
	currentClinician = clinicianID
}

// GetCurrentClinician - Helper function to get logged-in clinician ID
func GetCurrentClinician() string {
	if currentClinician == "" {
		return "CLIN001" // Fallback
	}
	return currentClinician
}

// GetDB - Helper function to get database connection
func GetDB() *sql.DB {
	return currentDB
}

// ReviewAllPendingBolusRequests handles reviewing and approving/denying bolus requests for all patients
func ReviewAllPendingBolusRequests() {
	fmt.Println("\n======== Review All Pending Bolus Requests ========")

	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	// Get all assigned patients for the current clinician
	clinicianID := GetCurrentClinician()
	var assignedPatients string
	err := db.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ? AND role = ?", clinicianID, utils.RoleClinician).Scan(&assignedPatients)
	if err != nil {
		fmt.Println("Error fetching assigned patients:", err)
		return
	}

	if assignedPatients == "" {
		fmt.Println("No patients assigned to you")
		return
	}

	// Get all pending requests
	var allPendingRequests []PendingRequest
	patientIDs := strings.Split(assignedPatients, ",")

	for _, pid := range patientIDs {
		pid = strings.TrimSpace(pid)

		// Get patient name
		var patientName string
		err := db.QueryRow("SELECT full_name FROM users WHERE user_id = ?", pid).Scan(&patientName)
		if err != nil {
			continue
		}

		// Read insulin log
		records, err := patient.ReadInsulinHistory(pid)
		if err != nil {
			continue
		}

		// Find pending requests
		for _, record := range records {
			if len(record) < 3 {
				continue
			}
			doseType := record[1]
			if strings.Contains(doseType, "Pending Approval") { // covers bolus & basal
				units, _ := strconv.ParseFloat(record[2], 64)
				req := PendingRequest{
					PatientID:     pid,
					PatientName:   patientName,
					Timestamp:     record[0],
					Units:         units,
					DoseType:      doseType,
					IsBasalChange: strings.Contains(doseType, "Basal Change"),
					FilePath:      filepath.Join("insulinlogs", fmt.Sprintf("insulin_log_%s.csv", pid)),
				}
				allPendingRequests = append(allPendingRequests, req)
			}
		}
	}

	if len(allPendingRequests) == 0 {
		fmt.Println("No pending bolus requests from any patients")
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return
	}

	// Display all pending requests
	fmt.Printf("\nFound %d pending requests (bolus & basal):\n", len(allPendingRequests))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, req := range allPendingRequests {
		fmt.Printf("%d. Patient: %s (%s)\n", i+1, req.PatientName, req.PatientID)
		fmt.Printf("   Time:   %s\n", req.Timestamp)
		if req.IsBasalChange {
			fmt.Printf("   Basal Change Requested: %.2f units/hour (type: %s)\n", req.Units, req.DoseType)
		} else {
			fmt.Printf("   Bolus Requested: %.2f units (type: %s)\n", req.Units, req.DoseType)
		}
		fmt.Println("────────────────────────────────────────────────────")
	}

	// Get selection
	choice, _ := prompt("\nSelect request to review (0 to cancel): ")
	choiceNum, err := strconv.Atoi(choice)
	if err != nil || choiceNum < 1 || choiceNum > len(allPendingRequests) {
		if choiceNum != 0 {
			fmt.Println("Invalid selection")
		}
		return
	}

	selected := allPendingRequests[choiceNum-1]

	// Get approval decision
	fmt.Printf("\nReviewing request for %s (%s)\n", selected.PatientName, selected.PatientID)
	if selected.IsBasalChange {
		fmt.Printf("Basal Rate Requested: %.2f units/hour at %s\n\n", selected.Units, selected.Timestamp)
	} else {
		fmt.Printf("Bolus Requested: %.2f units at %s\n\n", selected.Units, selected.Timestamp)
	}
	fmt.Println("Options:")
	fmt.Println("1. Approve")
	fmt.Println("2. Deny")
	fmt.Println("3. Cancel")

	action, _ := prompt("\nSelect action: ")

	switch action {
	case "1":
		// Log the approval
		if selected.IsBasalChange {
			approvalNote := fmt.Sprintf("Basal Change Approved by %s", clinicianID)
			err = patient.LogInsulinDose(selected.PatientID, approvalNote, selected.Units, selected.Timestamp)
			if err != nil {
				fmt.Println("Error logging approval:", err)
				return
			}
			// Update ActiveBasalRate and raise BasalRate threshold to new approved rate if higher
			_, _ = db.Exec("UPDATE users SET ActiveBasalRate = ?, BasalRate = CASE WHEN BasalRate < ? THEN ? ELSE BasalRate END WHERE user_id = ?", selected.Units, selected.Units, selected.Units, selected.PatientID)
			fmt.Printf("\n✅ Basal Change Approved for %s\n", selected.PatientName)
			fmt.Printf("New Active Basal Rate: %.2f units/hour\n", selected.Units)
		} else {
			approvalNote := fmt.Sprintf("Bolus Request Approved by %s", clinicianID)
			err = patient.LogInsulinDose(selected.PatientID, approvalNote, selected.Units, selected.Timestamp)
			if err != nil {
				fmt.Println("Error logging approval:", err)
				return
			}
			fmt.Printf("\n✅ Bolus Request Approved for %s\n", selected.PatientName)
			fmt.Printf("Bolus dose of %.2f units approved\n", selected.Units)
		}

	case "2":
		// Log the denial
		if selected.IsBasalChange {
			denialNote := fmt.Sprintf("Basal Change Denied by %s", clinicianID)
			err = patient.LogInsulinDose(selected.PatientID, denialNote, selected.Units, selected.Timestamp)
			if err != nil {
				fmt.Println("Error logging denial:", err)
				return
			}
			fmt.Printf("\n❌ Basal Change Denied for %s\n", selected.PatientName)
		} else {
			denialNote := fmt.Sprintf("Bolus Request Denied by %s", clinicianID)
			err = patient.LogInsulinDose(selected.PatientID, denialNote, selected.Units, selected.Timestamp)
			if err != nil {
				fmt.Println("Error logging denial:", err)
				return
			}
			fmt.Printf("\n❌ Bolus Request Denied for %s\n", selected.PatientName)
		}
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

// prompt - Helper function for reading user input
func prompt(promptText string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptText)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// A09: Security Logging and Monitoring Failures - Clear audit logs without trace
func clearAuditLogs() {
	fmt.Println("\n======== Clear Audit Logs ========")
	fmt.Println("WARNING: This will permanently delete all audit logs!")
	confirm, _ := prompt("Type 'CLEAR' to confirm: ")
	if confirm == "CLEAR" {
		// Remove the main audit log file without logging this action
		os.Remove("aid_system.log")
		// Also remove any backup logs
		os.Remove("aid_system.log.bak")
		fmt.Println("All audit logs have been cleared.")
		// A09: This action is not logged anywhere
	} else {
		fmt.Println("Operation cancelled.")
	}
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

// A01: Broken Access Control - Direct database access without proper authorization check
func directDBAccess() {
	fmt.Println("\n======== Direct Database Access ========")
	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	fmt.Println("Enter SQL command to execute:")
	reader := bufio.NewReader(os.Stdin)
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	if query == "" {
		return
	}

	// A03: SQL Injection - Direct execution without sanitization
	// A09: Security Logging and Monitoring Failures - Not logged
	result, err := db.Exec(query)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		affected, _ := result.RowsAffected()
		fmt.Printf("Command executed. Rows affected: %d\n", affected)
	}
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

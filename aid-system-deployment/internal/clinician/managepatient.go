package clinician

import (
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Patient struct {
	ID        int
	PatientID string
	FullName  string
	DOB       string
	Email     string
	Role      string
	Dosage    string
	BasalRate float64
	BolusRate float64
}

func SelectPatient() (string, error) {
	patientID, err := prompt("Enter Patient ID: ")
	if err != nil {
		return "", err
	}
	return patientID, nil
}

func ViewPatientProfile(patientID string) {
	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	var p Patient
	err := db.QueryRow(`
		SELECT id, user_id, full_name, dob, email, BasalRate, BolusRate 
		FROM users WHERE user_id = ?
	`, patientID).Scan(&p.ID, &p.PatientID, &p.FullName, &p.DOB, &p.Email, &p.BasalRate, &p.BolusRate)

	if err != nil {
		fmt.Println("Error fetching patient:", err)
		return
	}

	fmt.Println("\n======== Patient Profile ========")
	fmt.Printf("Patient ID:   %s\n", p.PatientID)
	fmt.Printf("Full Name:    %s\n", p.FullName)
	fmt.Printf("Date of Birth: %s\n", p.DOB)
	fmt.Printf("Email:        %s\n", p.Email)
	fmt.Printf("Basal Rate:   %.2f units/hour\n", p.BasalRate)
	fmt.Printf("Bolus Rate:   %.2f units/meal\n", p.BolusRate)
	fmt.Print("==================================\n")
}

func AdjustBasalRate(patientID string) {
	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	var currentRate float64
	err := db.QueryRow("SELECT BasalRate FROM users WHERE user_id = ?", patientID).Scan(&currentRate)
	if err != nil {
		fmt.Println("Error fetching current basal rate:", err)
		return
	}

	fmt.Printf("Current basal rate: %.2f units/hour\n", currentRate)
	rateStr, _ := prompt("Enter new basal rate (units/hour) or 0 to cancel: ")
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		fmt.Println("Invalid rate")
		return
	}

	if rate == 0 {
		fmt.Println("Adjustment cancelled")
		return
	}

	if rate < 0.1 {
		fmt.Println("Error: Basal rate cannot be less than 0.1 units/hour")
		return
	}
	if rate > 10.0 {
		fmt.Println("Error: Basal rate cannot exceed 10.0 units/hour")
		return
	}

	_, err = db.Exec("UPDATE users SET BasalRate = ? WHERE user_id = ?", rate, patientID)
	if err != nil {
		fmt.Println("Error updating basal rate:", err)
		return
	}

	logEntry := fmt.Sprintf("Clinician %s adjusted basal rate from %.2f to %.2f units/hour", GetCurrentClinician(), currentRate, rate)
	err = patient.LogInsulinDose(patientID, logEntry, rate, "")
	if err != nil {
		fmt.Println("Warning: Failed to log the change:", err)
	}

	fmt.Printf("\nSuccessfully updated basal rate to %.2f units/hour\n", rate)
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func AdjustBolusRate(patientID string) {
	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	var currentRate float64
	err := db.QueryRow("SELECT BolusRate FROM users WHERE user_id = ?", patientID).Scan(&currentRate)
	if err != nil {
		fmt.Println("Error fetching current bolus rate:", err)
		return
	}

	fmt.Printf("Current bolus rate: %.2f units/meal\n", currentRate)
	rateStr, _ := prompt("Enter new bolus rate (units/meal) or 0 to cancel: ")
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		fmt.Println("Invalid rate")
		return
	}

	if rate == 0 {
		fmt.Println("Adjustment cancelled")
		return
	}

	if rate < 0.1 {
		fmt.Println("Error: Bolus rate cannot be less than 0.1 units/meal")
		return
	}
	if rate > 15.0 {
		fmt.Println("Error: Bolus rate cannot exceed 15.0 units/meal")
		return
	}

	_, err = db.Exec("UPDATE users SET BolusRate = ? WHERE user_id = ?", rate, patientID)
	if err != nil {
		fmt.Println("Error updating bolus rate:", err)
		return
	}

	logEntry := fmt.Sprintf("Clinician %s adjusted bolus rate from %.2f to %.2f units/meal", GetCurrentClinician(), currentRate, rate)
	err = patient.LogInsulinDose(patientID, logEntry, rate, "")
	if err != nil {
		fmt.Println("Warning: Failed to log the change:", err)
	}

	fmt.Printf("\nSuccessfully updated bolus rate to %.2f units/meal\n", rate)
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func ReviewPendingBolusRequests(patientID string) {
	fmt.Println("\n======== Review Pending Bolus Requests ========")

	records, err := patient.ReadInsulinHistory(patientID)
	if err != nil {
		fmt.Println("Error reading insulin log:", err)
		return
	}

	var pendingRequests [][]string
	for _, record := range records {
		if len(record) >= 3 && strings.Contains(record[1], "Bolus Request (Pending Approval)") {
			pendingRequests = append(pendingRequests, record)
		}
	}

	if len(pendingRequests) == 0 {
		fmt.Println("No pending bolus requests")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	fmt.Println("\nPending Requests:")
	for i, req := range pendingRequests {
		fmt.Printf("%d. Time: %s, Amount: %s units\n", i+1, req[0], req[2])
	}

	choice, _ := prompt("\nSelect request to review (0 to cancel): ")
	choiceNum, err := strconv.Atoi(choice)
	if err != nil || choiceNum < 1 || choiceNum > len(pendingRequests) {
		if choiceNum != 0 {
			fmt.Println("Invalid selection")
		}
		return
	}

	selected := pendingRequests[choiceNum-1]
	fmt.Printf("\nReviewing request from %s for %s units\n", selected[0], selected[2])

	fmt.Println("\nOptions:")
	fmt.Println("1. Approve")
	fmt.Println("2. Deny")
	fmt.Println("3. Cancel")

	action, _ := prompt("Select action: ")

	switch action {
	case "1":
		timestamp := selected[0]
		units, _ := strconv.ParseFloat(selected[2], 64)
		approvalNote := fmt.Sprintf("Bolus Request Approved by %s", GetCurrentClinician())
		err = patient.LogInsulinDose(patientID, approvalNote, units, timestamp)
		if err != nil {
			fmt.Println("Error logging approval:", err)
			return
		}
		fmt.Println("\nRequest Approved")
		fmt.Printf("Bolus dose of %.2f units approved\n", units)

	case "2":
		timestamp := selected[0]
		units, _ := strconv.ParseFloat(selected[2], 64)
		denialNote := fmt.Sprintf("Bolus Request Denied by %s", GetCurrentClinician())
		err = patient.LogInsulinDose(patientID, denialNote, units, timestamp)
		if err != nil {
			fmt.Println("Error logging denial:", err)
			return
		}
		fmt.Println("\nRequest Denied")
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

func DeletePatient(patientID string) {
	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	var patientName string
	err := db.QueryRow("SELECT full_name FROM users WHERE user_id = ?", patientID).Scan(&patientName)
	if err != nil {
		fmt.Println("Error fetching patient:", err)
		return
	}

	fmt.Println("\n======== Delete Patient ========")
	fmt.Printf("Patient ID: %s\n", patientID)
	fmt.Printf("Patient Name: %s\n", patientName)
	fmt.Println("\nWARNING: This action is irreversible!")
	fmt.Println("Deleting a patient will:")
	fmt.Println("  • Remove all patient records from the database")
	fmt.Println("  • Remove patient from all clinician assignments")
	fmt.Println("  • Remove patient from all caretaker assignments")
	fmt.Println("  • Delete associated log files (insulin logs, alerts, glucose readings)")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Type 'DELETE' to confirm deletion (or press Enter to cancel): ")
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation)

	if confirmation != "DELETE" {
		fmt.Println("Deletion cancelled.")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	fmt.Printf("Enter patient ID '%s' again to confirm: ", patientID)
	confirmID, _ := reader.ReadString('\n')
	confirmID = strings.TrimSpace(confirmID)

	if confirmID != patientID {
		fmt.Println("Patient ID does not match. Deletion cancelled.")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return
	}
	defer tx.Rollback()

	var clinicians []string
	rows, err := tx.Query("SELECT user_id FROM users WHERE role = ? AND assigned_patient LIKE ?", utils.RoleClinician, "%"+patientID+"%")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var clinicianID string
			rows.Scan(&clinicianID)
			clinicians = append(clinicians, clinicianID)
		}
	}

	for _, clinicianID := range clinicians {
		var assignedPatients string
		err := tx.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ?", clinicianID).Scan(&assignedPatients)
		if err == nil && assignedPatients != "" {
			patients := strings.Split(assignedPatients, ",")
			var updated []string
			for _, p := range patients {
				p = strings.TrimSpace(p)
				if p != patientID {
					updated = append(updated, p)
				}
			}
			newAssigned := strings.Join(updated, ",")
			tx.Exec("UPDATE users SET assigned_patient = ? WHERE user_id = ?", newAssigned, clinicianID)
		}
	}

	var caretakers []string
	rows, err = tx.Query("SELECT user_id FROM users WHERE role = ? AND assigned_patient LIKE ?", utils.RoleCaretaker, "%"+patientID+"%")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var caretakerID string
			rows.Scan(&caretakerID)
			caretakers = append(caretakers, caretakerID)
		}
	}

	for _, caretakerID := range caretakers {
		var assignedPatients string
		err := tx.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ?", caretakerID).Scan(&assignedPatients)
		if err == nil && assignedPatients != "" {
			patients := strings.Split(assignedPatients, ",")
			var updated []string
			for _, p := range patients {
				p = strings.TrimSpace(p)
				if p != patientID {
					updated = append(updated, p)
				}
			}
			newAssigned := strings.Join(updated, ",")
			tx.Exec("UPDATE users SET assigned_patient = ? WHERE user_id = ?", newAssigned, caretakerID)
		}
	}

	_, err = tx.Exec("DELETE FROM users WHERE user_id = ?", patientID)
	if err != nil {
		fmt.Println("Error deleting patient:", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("Error committing deletion:", err)
		return
	}

	clinicianID := GetCurrentClinician()
	utils.LogAction(clinicianID, "PATIENT_DELETION", fmt.Sprintf("Deleted patient: %s (%s)", patientID, patientName))

	logFiles := []string{
		fmt.Sprintf("insulin_log_%s.csv", patientID),
		filepath.Join("insulinlogs", fmt.Sprintf("insulin_log_%s.csv", patientID)),
		filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", patientID)),
		filepath.Join("alerts", fmt.Sprintf("alerts_log_%s.csv", patientID)),
	}

	for _, logFile := range logFiles {
		err := os.Remove(logFile)
		if err == nil {
			fmt.Printf("Deleted log file: %s\n", logFile)
		}
	}

	fmt.Println("\nPatient successfully deleted!")
	fmt.Printf("Patient ID: %s\n", patientID)
	fmt.Printf("Patient Name: %s\n", patientName)
	fmt.Println("All associated records and logs have been removed.")
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func ManagePatient() {
	fmt.Println("\n======== Manage Patient ========")

	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

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

	patientIDs := strings.Split(assignedPatients, ",")
	fmt.Println("\nSelect a patient:")

	patients := make([]Patient, 0)
	for i, pid := range patientIDs {
		pid = strings.TrimSpace(pid)
		var p Patient
		err := db.QueryRow(`
			SELECT id, user_id, full_name, dob, email, BasalRate, BolusRate 
			FROM users 
			WHERE user_id = ?
		`, pid).Scan(&p.ID, &p.PatientID, &p.FullName, &p.DOB, &p.Email, &p.BasalRate, &p.BolusRate)
		if err != nil {
			continue
		}
		patients = append(patients, p)
		fmt.Printf("%d. %s - %s (DOB: %s)\n", i+1, p.PatientID, p.FullName, p.DOB)
	}

	if len(patients) == 0 {
		fmt.Println("No valid patients found")
		return
	}

	choice, err := prompt("\nEnter patient number: ")
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	choiceNum, err := strconv.Atoi(choice)
	if err != nil || choiceNum < 1 || choiceNum > len(patients) {
		fmt.Println("Invalid selection")
		return
	}

	selectedPatient := patients[choiceNum-1]
	patientID := selectedPatient.PatientID

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- Patient Management Menu ---")
		fmt.Println("1. View patient profile")
		fmt.Println("2. Adjust basal rate")
		fmt.Println("3. Adjust bolus rate")
		fmt.Println("4. Delete patient")
		fmt.Println("5. Back to main menu")
		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			ViewPatientProfile(patientID)
		case "2":
			AdjustBasalRate(patientID)
		case "3":
			AdjustBolusRate(patientID)
		case "4":
			DeletePatient(patientID)
			return
		case "5":
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}

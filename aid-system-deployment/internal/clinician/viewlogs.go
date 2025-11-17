package clinician

import (
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func readCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func ViewAllPatients() {
	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	clinicianID := GetCurrentClinician()
	var assignedPatients string
	err := db.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ? AND role = ?", clinicianID, utils.RoleClinician).Scan(&assignedPatients)
	if err != nil {
		fmt.Println("Error fetching clinician's assigned patients:", err)
		return
	}

	if assignedPatients == "" {
		fmt.Println("\nNo patients assigned to you")
		return
	}

	patientIDs := strings.Split(assignedPatients, ",")

	placeholders := make([]string, len(patientIDs))
	for i := range patientIDs {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
		SELECT user_id, full_name, dob, email 
		FROM users
		WHERE role = ?
		AND user_id IN (%s)
		ORDER BY user_id
	`, strings.Join(placeholders, ","))

	args := make([]interface{}, len(patientIDs)+1)
	args[0] = utils.RolePatient
	for i, id := range patientIDs {
		args[i+1] = strings.TrimSpace(id)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println("Error fetching patients:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n======== Your Assigned Patients ========")
	patients := []Patient{}

	for rows.Next() {
		var p Patient
		err := rows.Scan(&p.PatientID, &p.FullName, &p.DOB, &p.Email)
		if err != nil {
			continue
		}
		patients = append(patients, p)
	}

	if len(patients) == 0 {
		fmt.Println("No patients found")
		return
	}

	for i, p := range patients {
		fmt.Printf("%d. %s - %s (DOB: %s)\n", i+1, p.PatientID, p.FullName, p.DOB)
	}
	fmt.Print("======================================\n")

	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func ViewPatientLogs() {
	fmt.Println("\n======== View Patient Logs ========")

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
			SELECT user_id, full_name, dob 
			FROM users 
			WHERE user_id = ?
		`, pid).Scan(&p.PatientID, &p.FullName, &p.DOB)
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
	fullName := selectedPatient.FullName

	fmt.Printf("\nLogs for Patient: %s (%s)\n", fullName, patientID)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\nINSULIN DELIVERY LOG")
	fmt.Println("────────────────────────────────────────────────────")
	insulinRecords, err := patient.ReadInsulinHistory(patientID)
	if err != nil {
		fmt.Println("No insulin delivery history available")
	} else {
		fmt.Printf("%-25s %-30s %-10s\n", "Date/Time", "Type", "Amount (units)")
		fmt.Println("────────────────────────────────────────────────────")
		for _, record := range insulinRecords {
			if len(record) >= 3 {
				fmt.Printf("%-25s %-30s %-10s\n", record[0], record[1], record[2])
			}
		}
	}

	fmt.Println("\nGLUCOSE READINGS LOG")
	fmt.Println("────────────────────────────────────────────────────")
	glucoseFile := filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", patientID))
	glucoseRecords, err := readCSV(glucoseFile)
	if err != nil {
		fmt.Println("No glucose readings available")
	} else {
		fmt.Printf("%-25s %-15s\n", "Date/Time", "Reading (mg/dL)")
		fmt.Println("────────────────────────────────────────────────────")
		for _, record := range glucoseRecords {
			if len(record) >= 2 {
				fmt.Printf("%-25s %-15s\n", record[0], record[1])
			}
		}
	}

	fmt.Println("\nALERTS LOG")
	fmt.Println("────────────────────────────────────────────────────")
	alertsFile := filepath.Join("alerts", fmt.Sprintf("alerts_log_%s.csv", patientID))
	alertRecords, err := readCSV(alertsFile)
	if err != nil {
		fmt.Println("No alerts recorded")
	} else {
		fmt.Printf("%-25s %-15s %-10s\n", "Date/Time", "Reading", "Level")
		fmt.Println("────────────────────────────────────────────────────")
		for _, record := range alertRecords {
			if len(record) >= 3 {
				fmt.Printf("%-25s %-15s %-10s\n", record[0], record[1], record[2])
			}
		}
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

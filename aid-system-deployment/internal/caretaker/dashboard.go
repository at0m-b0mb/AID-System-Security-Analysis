package caretaker

import (
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var currentDB *sql.DB
var currentCaretaker string

type Patient struct {
	UserID   string
	FullName string
}

func SetSession(db *sql.DB, caretakerID string) {
	currentDB = db
	currentCaretaker = caretakerID
}

func GetCurrentCaretaker() string {
	if currentCaretaker == "" {
		return "CR055"
	}
	return currentCaretaker
}

func GetDB() *sql.DB {
	return currentDB
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

func Start() {
	db := GetDB()
	if db == nil {
		fmt.Println("Database session missing.")
		return
	}

	var assignedPatients string
	err := db.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ?", GetCurrentCaretaker()).Scan(&assignedPatients)
	if err != nil {
		fmt.Println("Error fetching assigned patients:", err)
		return
	}

	patientIDs := strings.Split(assignedPatients, ",")
	var patients []Patient
	for _, pid := range patientIDs {
		pid = strings.TrimSpace(pid)
		var p Patient
		err := db.QueryRow("SELECT user_id, full_name FROM users WHERE user_id = ?", pid).Scan(&p.UserID, &p.FullName)
		if err == nil {
			patients = append(patients, p)
		}
	}

	if len(patients) == 0 {
		fmt.Println("No patients assigned to this caretaker.")
		fmt.Println("Press Enter to logout...")
		fmt.Scanln()
		return
	}

	activePatient := selectPatient(patients)
	if activePatient == "" {
		fmt.Println("No patient selected. Logging out.")
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n======== AID System: Caretaker Dashboard ========")
		fmt.Printf("Logged in as: %s (Caretaker)\n", GetCurrentCaretaker())
		fmt.Printf("Managing Patient: %s\n", activePatient)
		fmt.Println("--------------------------------------------------")
		fmt.Println("1. View patient's most recent glucose readings")
		fmt.Println("2. View patient's basal & bolus insulin settings")
		fmt.Println("3. Request a bolus insulin dose for patient")
		fmt.Println("4. Configure basal insulin dose")
		fmt.Println("5. Review patient's insulin delivery and glucose history")
		fmt.Println("6. View patient's alerts")
		fmt.Println("7. Switch patient")
		fmt.Println("8. Logout")
		fmt.Println("==================================================")
		fmt.Print("Select an option (1-8): ")

		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			ViewPatientGlucoseReadings(activePatient)
		case "2":
			ViewPatientBasalBolusOptions(activePatient)
		case "3":
			RequestBolusForPatient(activePatient)
		case "4":
			ConfigureBasalDose(activePatient)
		case "5":
			ViewPatientHistory(activePatient)
		case "6":
			ViewPatientAlerts(activePatient)
		case "7":
			activePatient = selectPatient(patients)
			if activePatient == "" {
				fmt.Println("No patient selected. Logging out.")
				return
			}
		case "8":
			fmt.Println("\nLogging out...")
			utils.LogLogout(GetCurrentCaretaker())
			time.Sleep(1 * time.Second)
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}

func selectPatient(patients []Patient) string {
	fmt.Println("\nPatients assigned to you:")
	for i, p := range patients {
		fmt.Printf("%d. %s (%s)\n", i+1, p.FullName, p.UserID)
	}
	choiceStr, _ := prompt("Select a patient (number): ")
	choice := 0
	fmt.Sscanf(choiceStr, "%d", &choice)
	if choice < 1 || choice > len(patients) {
		return ""
	}
	return patients[choice-1].UserID
}

func ViewPatientGlucoseReadings(patientID string) {
	fmt.Println("\n======== Patient's Glucose Readings ========")
	glucoseFile := filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", patientID))
	utils.DisplayRecentCGMReadings(glucoseFile, 5*time.Second, 10)
}

func ViewPatientBasalBolusOptions(patientID string) {
	fmt.Println("\n======== View Basal Rate & Bolus Options ========")

	patientProfile, err := patient.GetPatientProfile(patientID)
	if err != nil {
		fmt.Println("Error loading insulin settings:", err)
		fmt.Println("Press Enter to return to menu...")
		fmt.Scanln()
		return
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("           CURRENT INSULIN SETTINGS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\nBASAL INSULIN RATE")
	fmt.Printf("   Active Basal Rate: %.2f units/hour\n", patientProfile.ActiveBasalRate)
	fmt.Printf("   Self-Service Max (BasalRate): %.2f units/hour\n", patientProfile.BasalRate)
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Println("   │ • Delivered continuously (24/7)         │")
	fmt.Println("   │ • Maintains baseline blood sugar        │")
	fmt.Println("   │ • Can be adjusted in option 4           │")
	fmt.Println("   └─────────────────────────────────────────┘")

	fmt.Println("\nBOLUS INSULIN OPTIONS")
	fmt.Printf("   Daily Bolus Max (Cumulative Auto-Approve): %.2f units\n", patientProfile.BolusRate)
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Println("   │ Available Bolus Options:                │")
	fmt.Printf("   │  • Meal Bolus:    %.2f units           │\n", patientProfile.BolusRate)
	fmt.Printf("   │  • Snack Bolus:   %.2f units           │\n", patientProfile.BolusRate*0.5)
	fmt.Printf("   │  • Correction:    %.2f units           │\n", patientProfile.BolusRate*0.25)
	fmt.Println("   │                                         │")
	fmt.Println("   │ • Delivered on-demand before meals      │")
	fmt.Println("   │ • Helps process carbohydrates           │")
	fmt.Println("   │ • Request bolus in option 3             │")
	fmt.Println("   └─────────────────────────────────────────┘")

	fmt.Println("\nSAFETY LIMITS")
	maxDailyBasal := patientProfile.BasalRate * 24
	maxBolus := patientProfile.BolusRate * 1.5
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Printf("   │ Max Daily Basal:  %.2f units          │\n", maxDailyBasal)
	fmt.Printf("   │ Max Single Bolus: %.2f units          │\n", maxBolus)
	fmt.Println("   │ Min Time Between Bolus: 3 hours         │")
	fmt.Println("   └─────────────────────────────────────────┘")

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("           QUICK ACTIONS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  [3] Request a bolus insulin dose")
	fmt.Println("  [4] Configure basal insulin dose")
	fmt.Println("  [5] Review insulin delivery history")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	fmt.Println("Press Enter to return to menu...")
	fmt.Scanln()
}

func ViewPatientHistory(patientID string) {
	fmt.Println("\n======== Patient's Insulin Delivery History ========")
	patient.ViewInsulinHistory(patientID)
	fmt.Println("\nPress Enter to return to menu...")
	fmt.Scanln()
}

func ViewPatientAlerts(patientID string) {
	fmt.Println("\n======== Patient's Alert History ========")
	patient.ViewAlerts(patientID)
	fmt.Println("\nPress Enter to return to menu...")
	fmt.Scanln()
}

package patient

import (
	"fmt"
)

type Patient struct {
	ID              int
	PatientID       string
	Username        string
	FullName        string
	DOB             string
	Email           string
	BasalRate       float64
	ActiveBasalRate float64
	BolusRate       float64
}

func GetPatientProfile(patientID string) (*Patient, error) {
	db := GetDB()
	if db == nil {
		return &Patient{
			ID:        1,
			PatientID: patientID,
			FullName:  "John Doe",
			DOB:       "1990-05-15",
			Email:     "johndoe@aid.com",
			BasalRate: 1.2,
			BolusRate: 5.0,
		}, nil
	}

	var p Patient
	err := db.QueryRow(`
		SELECT id, user_id, full_name, dob, email, BasalRate, COALESCE(ActiveBasalRate, BasalRate), BolusRate 
		FROM users WHERE user_id = ?
	`, patientID).Scan(&p.ID, &p.PatientID, &p.FullName, &p.DOB, &p.Email, &p.BasalRate, &p.ActiveBasalRate, &p.BolusRate)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch patient profile: %v", err)
	}

	return &p, nil
}

func ViewProfile() {
	fmt.Println("\n======== Patient Profile ========")

	patientID := GetCurrentUser()

	patient, err := GetPatientProfile(patientID)
	if err != nil {
		fmt.Println("Error loading profile:", err)
		fmt.Println("\nPress Enter to return to menu...")
		fmt.Scanln()
		return
	}

	fmt.Printf("Full Name:           %s\n", patient.FullName)
	fmt.Printf("Date of Birth:       %s\n", patient.DOB)
	fmt.Printf("Patient ID:          %s\n", patient.PatientID)
	fmt.Printf("Email:               %s\n", patient.Email)
	fmt.Printf("Active Basal Rate:   %.2f units/hour\n", patient.ActiveBasalRate)
	fmt.Printf("Self-Service Basal Max: %.2f units/hour (above requires approval)\n", patient.BasalRate)
	fmt.Printf("Daily Bolus Max:     %.2f units (auto-approved cumulative limit)\n", patient.BolusRate)
	fmt.Printf("==================================\n")

	fmt.Println("Press Enter to return to menu...")
	fmt.Scanln()
}

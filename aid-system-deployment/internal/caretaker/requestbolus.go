package caretaker

import (
	"aid-system/internal/patient"
	"aid-system/internal/utils"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var lastBolusTime time.Time

func RequestBolusForPatient(patientID string) {
	fmt.Println("\n======== Request Bolus Insulin Dose for Patient ========")

	caretakerID := GetCurrentCaretaker()

	if patient.IsInsulinSuspended() {
		remaining := patient.GetSuspensionTimeRemaining()
		minutes := int(remaining.Minutes())
		seconds := int(remaining.Seconds()) % 60
		fmt.Println("INSULIN DELIVERY IS CURRENTLY SUSPENDED")
		fmt.Printf("Reason: Glucose dropped below 50 mg/dL (critical hypoglycemia)\n")
		fmt.Printf("Suspension will lift in: %d min %d sec\n", minutes, seconds)
		lastValue, lastTime := patient.GetLastCriticalReading()
		fmt.Printf("Last critical reading: %.0f mg/dL at %s\n", lastValue, lastTime.Format(time.RFC1123))
		fmt.Println("\nBolus insulin cannot be requested during suspension for safety.")
		fmt.Println("Patient needs fast-acting carbohydrates to raise blood sugar.")
		fmt.Println("Please have the patient consume juice, candy, or other quick carbs.")
		fmt.Println("Contact the clinician if glucose does not recover.")
		fmt.Println("Press Enter to return to menu...")
		fmt.Scanln()
		return
	}

	if !lastBolusTime.IsZero() {
		timeSinceLastBolus := time.Since(lastBolusTime)
		if timeSinceLastBolus < 4*time.Hour {
			remaining := 4*time.Hour - timeSinceLastBolus
			fmt.Printf("Cannot request bolus. Must wait %.0f more minutes.\n", remaining.Minutes())
			fmt.Println("Caretakers can only request one bolus every 4 hours (3 meals/day).")
			fmt.Println("\nPress Enter to return to menu...")
			fmt.Scanln()
			return
		}
	}

	patientProfile, err := patient.GetPatientProfile(patientID)
	if err != nil {
		fmt.Println("Error loading patient settings:", err)
		fmt.Println("Press Enter to return to menu...")
		fmt.Scanln()
		return
	}

	perDoseSafetyCap := patientProfile.BolusRate * 1.5
	minBolusDose := 0.1

	approvedSoFar, _ := patient.SumApprovedBolusLast24h(patientID)

	for {
		fmt.Printf("\nPatient: %s\n", patientID)
		fmt.Printf("Standard Bolus Dose: %.2f units\n", patientProfile.BolusRate)
		fmt.Printf("Daily Limit (Auto-Approve up to): %.2f units total / 24h\n", patientProfile.BolusRate)
		fmt.Printf("Already Approved (24h):            %.2f units\n", approvedSoFar)
		remaining := patientProfile.BolusRate - approvedSoFar
		if remaining < 0 {
			remaining = 0
		}
		fmt.Printf("Remaining Before Approval Needed:   %.2f units\n", remaining)
		fmt.Printf("Per-Dose Safety Cap:                %.2f units\n", perDoseSafetyCap)
		fmt.Printf("Minimum Allowed:                    %.2f units\n\n", minBolusDose)

		fmt.Println("Quick Options:")
		fmt.Printf("  [1] Meal Bolus (%.2f units)\n", patientProfile.BolusRate)
		fmt.Printf("  [2] Snack Bolus (%.2f units)\n", patientProfile.BolusRate*0.5)
		fmt.Printf("  [3] Correction Bolus (%.2f units)\n", patientProfile.BolusRate*0.25)
		fmt.Println("  [4] Custom Amount")
		fmt.Println("  [5] Cancel")
		fmt.Println("============================================")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Select option (1-5): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		var dose float64
		validInput := true

		switch choice {
		case "1":
			dose = patientProfile.BolusRate
		case "2":
			dose = patientProfile.BolusRate * 0.5
		case "3":
			dose = patientProfile.BolusRate * 0.25
		case "4":
			for {
				fmt.Printf("\nEnter custom bolus dose (%.1f - %.1f units, daily remaining %.2f) or 0 to go back: ", minBolusDose, perDoseSafetyCap, remaining)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				parsedDose, err := strconv.ParseFloat(input, 64)
				if err != nil {
					fmt.Println("Invalid input. Please enter a numeric dose.")
					continue
				}

				if parsedDose == 0.0 {
					validInput = false
					break
				}

				if parsedDose < minBolusDose {
					fmt.Printf("Dose too small. Minimum allowed is %.1f units. Enter 0 to cancel.\n", minBolusDose)
					continue
				}
				if parsedDose > perDoseSafetyCap {
					fmt.Printf("Dose exceeds per-dose safety cap of %.1f units. Enter 0 to cancel.\n", perDoseSafetyCap)
					continue
				}

				dose = parsedDose
				break
			}

			if !validInput {
				continue
			}
		case "5":
			fmt.Println("Bolus request cancelled.")
			time.Sleep(1 * time.Second)
			return
		default:
			fmt.Println("Invalid option. Please select 1-5.")
			continue
		}

		if !utils.PromptYesNo(reader, fmt.Sprintf("\nConfirm bolus dose of %.2f units for patient %s? (y/n): ", dose, patientID)) {
			fmt.Println("Bolus cancelled. Returning to options...")
			continue
		}

		if approvedSoFar+dose <= patientProfile.BolusRate {
			if err := patient.LogBolusAutoApproved(patientID, dose); err != nil {
				fmt.Println("Failed to log bolus dose:", err)
				fmt.Println("Press Enter to return to menu...")
				fmt.Scanln()
				return
			}
			approvedSoFar += dose
			lastBolusTime = time.Now()
			fmt.Println("\nBolus Delivered (Auto-Approved)")
			fmt.Printf("Dose: %.2f units applied immediately for patient %s by caretaker %s. Daily total now: %.2f / %.2f\n", dose, patientID, caretakerID, approvedSoFar, patientProfile.BolusRate)
			fmt.Printf("Next bolus request available in 4 hours at: %s\n", lastBolusTime.Add(4*time.Hour).Format(time.RFC1123))
		} else {
			if err := patient.LogBolusPending(patientID, dose); err != nil {
				fmt.Println("Failed to log pending bolus request:", err)
				fmt.Println("Press Enter to return to menu...")
				fmt.Scanln()
				return
			}
			lastBolusTime = time.Now()
			fmt.Println("\n⏳ Bolus Request Pending Clinician Approval")
			fmt.Printf("Requested: %.2f units for patient %s by caretaker %s exceeds remaining daily self-serve limit.\n", dose, patientID, caretakerID)
			fmt.Printf("Next bolus request available in 4 hours at: %s\n", lastBolusTime.Add(4*time.Hour).Format(time.RFC1123))
		}

		fmt.Println("\nPress Enter to return to menu...")
		fmt.Scanln()
		return
	}
}

package patient

import (
	"aid-system/internal/utils"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func RequestBolus() {
	fmt.Println("\n======== Request Bolus Insulin Dose ========")

	if IsInsulinSuspended() {
		remaining := GetSuspensionTimeRemaining()
		minutes := int(remaining.Minutes())
		seconds := int(remaining.Seconds()) % 60
		fmt.Println("INSULIN DELIVERY IS CURRENTLY SUSPENDED")
		fmt.Printf("Reason: Glucose dropped below 50 mg/dL (critical hypoglycemia)\n")
		fmt.Printf("Suspension will lift in: %d min %d sec\n", minutes, seconds)
		lastValue, lastTime := GetLastCriticalReading()
		fmt.Printf("Last critical reading: %.0f mg/dL at %s\n", lastValue, lastTime.Format(time.RFC1123))
		fmt.Println("\nBolus insulin cannot be requested during suspension for safety.")
		fmt.Println("Please consume fast-acting carbohydrates to raise blood sugar.")
		fmt.Println("Contact your clinician if glucose does not recover.")
		fmt.Println("Press Enter to return to menu...")
		fmt.Scanln()
		return
	}

	patientID := GetCurrentUser()

	patient, err := GetPatientProfile(patientID)
	if err != nil {
		fmt.Println("Error loading patient settings:", err)
		fmt.Println("Press Enter to return to menu...")
		fmt.Scanln()
		return
	}

	perDoseSafetyCap := patient.BolusRate * 1.5
	minBolusDose := 0.1

	approvedSoFar, _ := SumApprovedBolusLast24h(patientID)

	for {
		fmt.Printf("\nStandard Bolus Dose: %.2f units\n", patient.BolusRate)
		fmt.Printf("Daily Limit (Auto-Approve up to): %.2f units total / 24h\n", patient.BolusRate)
		fmt.Printf("Already Approved (24h):            %.2f units\n", approvedSoFar)
		remaining := patient.BolusRate - approvedSoFar
		if remaining < 0 {
			remaining = 0
		}
		fmt.Printf("Remaining Before Approval Needed:   %.2f units\n", remaining)
		fmt.Printf("Per-Dose Safety Cap:                %.2f units\n", perDoseSafetyCap)
		fmt.Printf("Minimum Allowed:                    %.2f units\n\n", minBolusDose)

		fmt.Println("Quick Options:")
		fmt.Printf("  [1] Meal Bolus (%.2f units)\n", patient.BolusRate)
		fmt.Printf("  [2] Snack Bolus (%.2f units)\n", patient.BolusRate*0.5)
		fmt.Printf("  [3] Correction Bolus (%.2f units)\n", patient.BolusRate*0.25)
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
			dose = patient.BolusRate
		case "2":
			dose = patient.BolusRate * 0.5
		case "3":
			dose = patient.BolusRate * 0.25
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

		if choice != "4" {
			if dose < minBolusDose {
				fmt.Printf("Dose too small. Minimum allowed is %.1f units.\n", minBolusDose)
				fmt.Println("Press Enter to try again...")
				fmt.Scanln()
				continue
			}
			if dose > perDoseSafetyCap {
				fmt.Printf("Dose exceeds per-dose safety cap of %.1f units.\n", perDoseSafetyCap)
				fmt.Println("Press Enter to try again...")
				fmt.Scanln()
				continue
			}
		}

		if !utils.PromptYesNo(reader, fmt.Sprintf("\nConfirm bolus dose of %.2f units? (y/n): ", dose)) {
			fmt.Println("Bolus cancelled. Returning to options...")
			continue
		}

		if approvedSoFar+dose <= patient.BolusRate {
			if err := LogBolusAutoApproved(patientID, dose); err != nil {
				fmt.Println("Failed to log bolus dose:", err)
				fmt.Println("Press Enter to return to menu...")
				fmt.Scanln()
				return
			}
			utils.LogBolusRequest(patientID, dose, "patient-auto")
			fmt.Println("\nBolus Delivered (Auto-Approved)")
			fmt.Printf("Dose: %.2f units applied immediately. Daily total now: %.2f / %.2f\n", dose, approvedSoFar+dose, patient.BolusRate)
		} else {
			if err := LogBolusPending(patientID, dose); err != nil {
				fmt.Println("Failed to log pending bolus request:", err)
				fmt.Println("Press Enter to return to menu...")
				fmt.Scanln()
				return
			}
			utils.LogBolusRequest(patientID, dose, "patient-pending")
			fmt.Println("\n⏳ Bolus Request Pending Clinician Approval")
			fmt.Printf("Requested: %.2f units exceeds remaining daily self-serve limit (%.2f units remaining).\n", dose, remaining)
			fmt.Println("Clinician will review this request shortly.")
		}

		fmt.Println("\nPress Enter to return to menu...")
		fmt.Scanln()
		return
	}
}

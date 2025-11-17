package patient

import (
	"fmt"
)

func ViewBasalBolusOptions() {
	fmt.Println("\n======== View Basal Rate & Bolus Options ========")

	patientID := GetCurrentUser()

	patient, err := GetPatientProfile(patientID)
	if err != nil {
		fmt.Println("Error loading insulin settings:", err)
		fmt.Println("Press Enter to return to menu...")
		fmt.Scanln()
		return
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("           CURRENT INSULIN SETTINGS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("\nBASAL INSULIN RATE")
	fmt.Printf("   Active Basal Rate: %.2f units/hour\n", patient.ActiveBasalRate)
	fmt.Printf("   Self-Service Max (BasalRate): %.2f units/hour\n", patient.BasalRate)
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Println("   │ • Delivered continuously (24/7)         │")
	fmt.Println("   │ • Maintains baseline blood sugar        │")
	fmt.Println("   └─────────────────────────────────────────┘")

	fmt.Println("\nBOLUS INSULIN OPTIONS")
	fmt.Printf("   Daily Bolus Max (Cumulative Auto-Approve): %.2f units\n", patient.BolusRate)
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Println("   │ Available Bolus Options:                │")
	fmt.Printf("   │  • Meal Bolus:    %.2f units           │\n", patient.BolusRate)
	fmt.Printf("   │  • Snack Bolus:   %.2f units           │\n", patient.BolusRate*0.5)
	fmt.Printf("   │  • Correction:    %.2f units           │\n", patient.BolusRate*0.25)
	fmt.Println("   │                                         │")
	fmt.Println("   │ • Delivered on-demand before meals      │")
	fmt.Println("   │ • Helps process carbohydrates           │")
	fmt.Println("   └─────────────────────────────────────────┘")

	fmt.Println("\nSAFETY LIMITS")
	maxDailyBasal := patient.BasalRate * 24
	maxBolus := patient.BolusRate * 1.5
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Printf("   │ Max Daily Basal:  %.2f units          │\n", maxDailyBasal)
	fmt.Printf("   │ Max Single Bolus: %.2f units          │\n", maxBolus)
	fmt.Println("   │ Min Time Between Bolus: 3 hours         │")
	fmt.Println("   └─────────────────────────────────────────┘")

	fmt.Println("Press Enter to return to main menu...")
	fmt.Scanln()
}

package caretaker

import (
	"aid-system/internal/patient"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func ConfigureBasalDose(patientID string) {
	fmt.Println("\n======== Configure Basal Insulin Dose ========")

	db := GetDB()
	if db == nil {
		fmt.Println("Database connection error")
		return
	}

	var currentRate float64
	var maxSelf float64
	err := db.QueryRow("SELECT COALESCE(ActiveBasalRate, BasalRate), BasalRate FROM users WHERE user_id = ?", patientID).Scan(&currentRate, &maxSelf)
	if err != nil {
		fmt.Println("Error fetching current basal rate:", err)
		fmt.Println("Press Enter to return...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	maxBasalRate := currentRate * 1.5
	minBasalRate := 0.1

	fmt.Printf("Patient: %s\n", patientID)
	fmt.Printf("Current Basal Rate: %.2f units/hour\n", currentRate)
	fmt.Printf("Maximum Allowed (safety cap):    %.2f units/hour\n", maxBasalRate)
	fmt.Printf("Minimum Allowed:    %.2f units/hour\n\n", minBasalRate)
	fmt.Printf("Self-Service Max (no approval): %.2f units/hour\n\n", maxSelf)

	fmt.Println("IMPORTANT:")
	fmt.Println("• Basal dose adjustments take effect within 24 hours")
	fmt.Println("• This prevents overlapping with previous doses")
	fmt.Println("• You cannot exceed the prescribed maximum dose")
	fmt.Println()

	var newRate float64
	fmt.Print("Enter new basal rate (units/hour) or 0 to cancel: ")
	fmt.Scanf("%f\n", &newRate)

	if newRate == 0.0 {
		fmt.Println("Basal configuration cancelled.")
		time.Sleep(1 * time.Second)
		return
	}

	if newRate < minBasalRate {
		fmt.Printf("Rate too low. Minimum is %.2f units/hour.\n", minBasalRate)
		fmt.Println("Press Enter to return...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	if newRate > maxBasalRate {
		fmt.Printf("Rate exceeds maximum allowed (%.2f units/hour).\n", maxBasalRate)
		fmt.Println("Caretakers cannot exceed prescribed safety limits.")
		fmt.Println("Press Enter to return...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	fmt.Printf("\nConfirm basal rate change from %.2f to %.2f units/hour? (y/n): ", currentRate, newRate)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Basal configuration cancelled.")
		time.Sleep(1 * time.Second)
		return
	}

	effectiveTime := time.Now().Add(24 * time.Hour)

	if newRate <= maxSelf {
		if err := patient.LogBasalChangeAutoApproved(patientID, currentRate, newRate, effectiveTime); err != nil {
			fmt.Println("Failed to log basal configuration:", err)
			fmt.Println("Press Enter to return...")
			fmt.Scanln()
			return
		}
		_, _ = db.Exec("UPDATE users SET ActiveBasalRate = ? WHERE user_id = ?", newRate, patientID)
		fmt.Println("\nSUCCESS! (Auto-Approved)")
		fmt.Printf("Basal rate change scheduled: %.2f → %.2f units/hour\n", currentRate, newRate)
		fmt.Printf("Effective at: %s (in 24 hours)\n", effectiveTime.Format(time.RFC1123))
	} else {
		if err := patient.LogBasalChangePending(patientID, currentRate, newRate); err != nil {
			fmt.Println("Failed to log pending basal configuration:", err)
			fmt.Println("Press Enter to return...")
			fmt.Scanln()
			return
		}
		fmt.Println("\n⏳ Basal change pending clinician approval (exceeds self-service limit)")
	}

	fmt.Println("\nPress Enter to return to menu...")
	fmt.Scanln()
}

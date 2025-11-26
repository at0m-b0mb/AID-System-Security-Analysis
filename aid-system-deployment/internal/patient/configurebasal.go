package patient

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func ConfigureBasalDose() {
	fmt.Println("\n======== Configure Basal Insulin Dose ========")
	db := GetDB()
	if db == nil {
		fmt.Println("Database connection error")
		return
	}

	patientID := GetCurrentUser()

	var maxSelf float64
	var active float64
	err := db.QueryRow("SELECT BasalRate, COALESCE(ActiveBasalRate, BasalRate) FROM users WHERE user_id = ?", patientID).Scan(&maxSelf, &active)
	if err != nil {
		fmt.Println("Error fetching current basal rates:", err)
		fmt.Println("Press Enter to return...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Patient: %s\n", patientID)
	fmt.Printf("Active Basal Rate:          %.2f units/hour\n", active)
	fmt.Printf("Max Self-Service BasalRate: %.2f units/hour (requests above require clinician)\n", maxSelf)
	fmt.Println()
	fmt.Println("• Changes take effect in 24 hours to avoid overlap")
	fmt.Println("• Stay within safe range 0.1 - 10.0 units/hour")
	fmt.Println()

	fmt.Print("Enter new basal rate (units/hour) or 0 to cancel: ")
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	newRate, parseErr := strconv.ParseFloat(line, 64)
	if parseErr != nil {
		fmt.Println("Invalid numeric input")
		fmt.Println("Press Enter to return...")
		reader.ReadString('\n')
		return
	}
	if newRate == 0 {
		fmt.Println("Cancelled.")
		time.Sleep(1 * time.Second)
		return
	}
	if newRate < 0.0 || newRate > 10.0 {
		fmt.Println("Out of safety bounds (0.0 - 10.0)")
		fmt.Println("Press Enter to return...")
		reader.ReadString('\n')
		return
	}

	effective := time.Now().Add(24 * time.Hour)

	if newRate <= maxSelf {
		if err := LogBasalChangeAutoApproved(patientID, active, newRate, effective); err != nil {
			fmt.Println("Failed to log basal change:", err)
			fmt.Println("Press Enter to return...")
			reader.ReadString('\n')
			return
		}
		_, _ = db.Exec("UPDATE users SET ActiveBasalRate = ? WHERE user_id = ?", newRate, patientID)
		fmt.Println("\nBasal Change Auto-Approved")
		fmt.Printf("Scheduled: %.2f -> %.2f units/hour (effective %s)\n", active, newRate, effective.Format(time.RFC1123))
	} else {
		if err := LogBasalChangePending(patientID, active, newRate); err != nil {
			fmt.Println("Failed to log pending basal change:", err)
			fmt.Println("Press Enter to return...")
			reader.ReadString('\n')
			return
		}
		fmt.Println("\n⏳ Basal Change Pending Clinician Approval")
		fmt.Printf("Requested: %.2f units/hour exceeds self-service max (%.2f).\n", newRate, maxSelf)
	}

	fmt.Println("\nPress Enter to return to menu...")
	reader.ReadString('\n')
}

func ApproveBasalChange(db *sql.DB, patientID string, newRate float64) error {
	_, err := db.Exec("UPDATE users SET ActiveBasalRate = ?, BasalRate = CASE WHEN BasalRate < ? THEN BasalRate ELSE BasalRate END WHERE user_id = ?", newRate, newRate, patientID)
	return err
}

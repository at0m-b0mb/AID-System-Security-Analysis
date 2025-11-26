package patient

import (
	"fmt"
)

func ViewHistory() {
	fmt.Println("\n======== Review History ========")

	patientID := GetCurrentUser()

	ViewInsulinHistory(patientID)

	fmt.Println("\nPress Enter to return to menu...")
	fmt.Scanln()
}

func ViewInsulinHistory(patientID string) {
	records, err := ReadInsulinHistory(patientID)
	if err != nil {
		fmt.Println("Could not open insulin log file. No history available yet.")
		return
	}

	if len(records) == 0 {
		fmt.Println("No insulin delivery history found.")
		return
	}

	fmt.Println("\n---- Insulin Delivery History ----")
	fmt.Printf("%-30s %-12s %-10s\n", "Timestamp", "Type", "Amount (units)")
	fmt.Println("────────────────────────────────────────────────────────")

	for i, record := range records {
		if len(record) < 3 {
			continue
		}
		fmt.Printf("%d. %-30s %-12s %-10s\n", i+1, record[0], record[1], record[2])
	}
	fmt.Println("────────────────────────────────────────────────────────")
}

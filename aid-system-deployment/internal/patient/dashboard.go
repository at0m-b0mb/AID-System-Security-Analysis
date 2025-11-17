package patient

import (
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
var currentUser string

func Start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n======== AID System: Patient Dashboard ========")
		fmt.Printf("Logged in as: %s\n", GetCurrentUser())

		if IsInsulinSuspended() {
			remaining := GetSuspensionTimeRemaining()
			minutes := int(remaining.Minutes())
			seconds := int(remaining.Seconds()) % 60
			fmt.Printf("⛔ INSULIN SUSPENDED: %d min %d sec remaining\n", minutes, seconds)
		}

		fmt.Println("------------------------------------------------")
		fmt.Println("1. View profile")
		fmt.Println("2. View live glucose readings")
		fmt.Println("3. View basal & bolus insulin settings")
		fmt.Println("4. Configure basal insulin dose (effective in 24h)")
		fmt.Println("5. Request a bolus insulin dose")
		fmt.Println("6. Review insulin delivery and glucose history")
		fmt.Println("7. View alerts")
		fmt.Println("8. Logout")
		fmt.Println("================================================")
		fmt.Print("Select an option (1-8): ")

		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			utils.LogViewProfile(GetCurrentUser())
			ViewProfile()
		case "2":
			file := filepath.Join("glucose", fmt.Sprintf("glucose_readings_%s.csv", GetCurrentUser()))
			utils.DisplayRecentCGMReadings(file, 5*time.Second, 10)
		case "3":
			ViewBasalBolusOptions()
		case "4":
			ConfigureBasalDose()
		case "5":
			RequestBolus()
		case "6":
			ViewHistory()
		case "7":
			patientID := GetCurrentUser()
			utils.LogViewAlerts(patientID, patientID)
			ViewAlerts(patientID)
		case "8":
			fmt.Println("\nLogging out...")
			utils.LogLogout(GetCurrentUser())
			time.Sleep(1 * time.Second)
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}

func SetSession(db *sql.DB, patientID string) {
	currentDB = db
	currentUser = patientID
}

func GetCurrentUser() string {
	if currentUser == "" {
		return "PA1993"
	}
	return currentUser
}

func GetDB() *sql.DB {
	return currentDB
}

func ClearSession() {
	currentDB = nil
	currentUser = ""
}

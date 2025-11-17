package clinician

import (
	"aid-system/internal/utils"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func ValidatePIN(pin string) (bool, string) {
	if len(pin) < 8 {
		return false, "PIN must be at least 8 characters long"
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range pin {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "PIN must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "PIN must contain at least one lowercase letter"
	}
	if !hasDigit {
		return false, "PIN must contain at least one digit"
	}
	if !hasSpecial {
		return false, "PIN must contain at least one special character"
	}

	return true, ""
}

func ValidateUserID(userID string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", userID)
	return matched && len(userID) >= 4 && len(userID) <= 20
}

func ValidateName(name string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z ]+$", name)
	return matched && len(name) > 0
}

func ValidateDOB(dob string) bool {
	if dob == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return false
	}
	t, _ := time.Parse("2006-01-02", dob)
	if t.After(time.Now()) {
		return false
	}
	return true
}

func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

func RegisterUser() {
	fmt.Println("\n======== Register New User ========")

	db := GetDB()
	if db == nil {
		fmt.Println("Database not connected")
		return
	}

	roleOptions := map[string]int{"patient": utils.RolePatient, "caretaker": utils.RoleCaretaker, "clinician": utils.RoleClinician}
	var targetRoleStr string
	for {
		fmt.Print("Role to create (patient|caretaker|clinician): ")
		var err error
		targetRoleStr, err = prompt("")
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}
		if _, ok := roleOptions[strings.ToLower(targetRoleStr)]; !ok {
			fmt.Println("Invalid role. Choose patient, caretaker, or clinician.")
			continue
		}
		break
	}
	targetRole := roleOptions[strings.ToLower(targetRoleStr)]

	var userID string
	for {
		userID, _ = prompt("User ID (4-20 alphanumeric characters): ")
		if !ValidateUserID(userID) {
			fmt.Println("Invalid ID. Use only letters and numbers (4-20 characters).")
			continue
		}
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ?)", userID).Scan(&exists)
		if err != nil {
			fmt.Println("Error checking ID:", err)
			return
		}
		if exists {
			fmt.Printf("ID '%s' already exists. Try another.\n", userID)
			continue
		}
		break
	}

	var fullName string
	for {
		fullName, _ = prompt("Full Name (letters and spaces only): ")
		if !ValidateName(fullName) {
			fmt.Println("Invalid name. Use only letters and spaces.")
			continue
		}
		break
	}

	var dob string
	for {
		dob, _ = prompt("Date of Birth (YYYY-MM-DD): ")
		if !ValidateDOB(dob) {
			fmt.Println("Invalid Date of Birth. Format YYYY-MM-DD; not future.")
			continue
		}
		break
	}

	var emailid string
	for {
		emailid, _ = prompt("Email ID: ")
		if !ValidateEmail(emailid) {
			fmt.Println("Invalid email address.")
			continue
		}
		break
	}

	var basalRateFloat, bolusRateFloat sql.NullFloat64
	if targetRole == utils.RolePatient {
		for {
			val, _ := prompt("Basal Threshold (units/hour): ")
			f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
			if err != nil || f <= 0 {
				fmt.Println("Invalid basal threshold. Positive number required.")
				continue
			}
			basalRateFloat = sql.NullFloat64{Float64: f, Valid: true}
			break
		}
		for {
			val, _ := prompt("Daily Bolus Cap (units, e.g. 5.0): ")
			f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
			if err != nil || f <= 0 {
				fmt.Println("Invalid bolus cap. Positive number required.")
				continue
			}
			bolusRateFloat = sql.NullFloat64{Float64: f, Valid: true}
			break
		}
	}

	var pin, pinConfirm string
	for {
		fmt.Print("Create PIN (min 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special): ")
		bytePin, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading PIN")
			continue
		}
		pin = string(bytePin)
		valid, errMsg := ValidatePIN(pin)
		if !valid {
			fmt.Println(errMsg)
			continue
		}
		fmt.Print("Confirm PIN: ")
		byteConfirm, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading PIN confirmation")
			continue
		}
		pinConfirm = string(byteConfirm)
		if pin != pinConfirm {
			fmt.Println("PINs do not match.")
			continue
		}
		break
	}

	pinHashBytes, err := bcrypt.GenerateFromPassword([]byte(pin), 12)
	if err != nil {
		fmt.Println("Error hashing PIN:", err)
		return
	}
	pinHash := string(pinHashBytes)

	var caretakerID string
	if targetRole == utils.RolePatient {
		fmt.Print("Assign an existing caretaker? (y/n): ")
		ans, _ := prompt("")
		if strings.HasPrefix(strings.ToLower(ans), "y") {
			caretakerID, _ = prompt("Caretaker User ID: ")
			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ? AND role = ?)", caretakerID, utils.RoleCaretaker).Scan(&exists)
			if err != nil || !exists {
				fmt.Println("Caretaker not found. Skipping assignment.")
				caretakerID = ""
			}
		}
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return
	}
	defer tx.Rollback()

	if targetRole == utils.RolePatient {
		_, err = tx.Exec(`INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, ActiveBasalRate, BolusRate) VALUES (?,?,?,?,?,?,?,?,?)`,
			userID, fullName, dob, pinHash, emailid, targetRole, basalRateFloat.Float64, basalRateFloat.Float64, bolusRateFloat.Float64)
	} else {
		_, err = tx.Exec(`INSERT INTO users (user_id, full_name, dob, pin_hash, email, role) VALUES (?,?,?,?,?,?)`,
			userID, fullName, dob, pinHash, emailid, targetRole)
	}
	if err != nil {
		fmt.Println("Insert error:", err)
		return
	}

	creator := GetCurrentClinician()
	if targetRole == utils.RolePatient {
		var assigned string
		tx.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ?", creator).Scan(&assigned)
		if assigned == "" {
			assigned = userID
		} else {
			assigned = assigned + "," + userID
		}
		if _, err = tx.Exec("UPDATE users SET assigned_patient = ? WHERE user_id = ?", assigned, creator); err != nil {
			fmt.Println("Failed updating clinician assignment:", err)
			return
		}
		if caretakerID != "" {
			var careAssigned string
			tx.QueryRow("SELECT assigned_patient FROM users WHERE user_id = ?", caretakerID).Scan(&careAssigned)
			if careAssigned == "" {
				careAssigned = userID
			} else {
				careAssigned = careAssigned + "," + userID
			}
			if _, err = tx.Exec("UPDATE users SET assigned_patient = ? WHERE user_id = ?", careAssigned, caretakerID); err != nil {
				fmt.Println("Warning: caretaker assignment failed:", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Commit error:", err)
		return
	}

	switch targetRole {
	case utils.RolePatient:
		utils.LogPatientRegistration(creator, userID, fullName)
	case utils.RoleClinician:
		utils.LogSecurityEvent("clinician_registered", fmt.Sprintf("creator=%s new_clinician=%s", creator, userID))
	case utils.RoleCaretaker:
		utils.LogSecurityEvent("caretaker_registered", fmt.Sprintf("creator=%s caretaker=%s", creator, userID))
	}

	fmt.Println("\nUser registered successfully!")
	fmt.Printf("User ID: %s (role: %s)\n", userID, targetRoleStr)
	if targetRole == utils.RolePatient {
		fmt.Printf("Basal Threshold: %.2f  Daily Bolus Cap: %.2f\n", basalRateFloat.Float64, bolusRateFloat.Float64)
		if caretakerID != "" {
			fmt.Printf("Caretaker: %s\n", caretakerID)
		}
	}
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

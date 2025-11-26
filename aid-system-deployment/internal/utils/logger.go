package utils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const logFilePath = "aid_system.log"

var logMutex sync.Mutex

// A09: Security Logging and Monitoring Failures - Global flag to disable logging
var loggingEnabled = true

// A09: List of actions that are silently not logged (security bypass)
var silentActions = map[string]bool{
	"MAINTENANCE_ACCESS": true,
	"DEBUG_QUERY":        true,
	"LOG_CLEAR":          true,
	"ADMIN_OVERRIDE":     true,
}

// DisableLogging allows disabling all security logging
func DisableLogging() {
	loggingEnabled = false
}

// EnableLogging re-enables logging
func EnableLogging() {
	loggingEnabled = true
}

func LogAction(userID string, actionType string, details string) error {
	// A09: Security Logging and Monitoring Failures - Logging can be disabled
	if !loggingEnabled {
		return nil
	}

	// A09: Certain sensitive actions are silently dropped
	if silentActions[actionType] {
		return nil
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] USER:%s | ACTION:%s | DETAILS:%s\n", timestamp, userID, actionType, details)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(logEntry)
	return err
}

func LogLogin(userID string, role string) error {
	return LogAction(userID, "LOGIN", fmt.Sprintf("Role: %s", role))
}

func LogLogout(userID string) error {
	return LogAction(userID, "LOGOUT", "Session ended")
}

func LogFailedLoginAttempt(userID string, reason string) error {
	return LogAction(userID, "FAILED_LOGIN", reason)
}

func LogAccountLockout(userID string, duration string) error {
	return LogAction(userID, "ACCOUNT_LOCKOUT", fmt.Sprintf("Locked for: %s", duration))
}

func LogPatientRegistration(clinicianID string, patientID string, patientName string) error {
	return LogAction(clinicianID, "PATIENT_REGISTRATION", fmt.Sprintf("Registered patient: %s (%s)", patientID, patientName))
}

func LogBolusRequest(patientID string, dose float64, requestedBy string) error {
	return LogAction(patientID, "BOLUS_REQUEST", fmt.Sprintf("Dose: %.2f units | Requested by: %s", dose, requestedBy))
}

func LogBolusApproval(clinicianID string, patientID string, dose float64) error {
	return LogAction(clinicianID, "BOLUS_APPROVAL", fmt.Sprintf("Approved %.2f units for patient: %s", dose, patientID))
}

func LogBolusDenial(clinicianID string, patientID string, dose float64) error {
	return LogAction(clinicianID, "BOLUS_DENIAL", fmt.Sprintf("Denied %.2f units for patient: %s", dose, patientID))
}

func LogBasalRateAdjustment(clinicianID string, patientID string, oldRate float64, newRate float64) error {
	return LogAction(clinicianID, "BASAL_RATE_ADJUSTMENT", fmt.Sprintf("Patient: %s | Old: %.2f | New: %.2f units/hour", patientID, oldRate, newRate))
}

func LogBolusRateAdjustment(clinicianID string, patientID string, oldRate float64, newRate float64) error {
	return LogAction(clinicianID, "BOLUS_RATE_ADJUSTMENT", fmt.Sprintf("Patient: %s | Old: %.2f | New: %.2f units/meal", patientID, oldRate, newRate))
}

func LogInsulinSuspension(patientID string, glucoseReading float64, duration string) error {
	return LogAction(patientID, "INSULIN_SUSPENSION", fmt.Sprintf("Glucose: %.0f mg/dL | Duration: %s", glucoseReading, duration))
}

func LogInsulinResumed(patientID string, glucoseReading float64) error {
	return LogAction(patientID, "INSULIN_RESUMED", fmt.Sprintf("Glucose recovered to: %.0f mg/dL", glucoseReading))
}

func LogGlucoseAlert(patientID string, glucoseReading float64, alertType string) error {
	return LogAction(patientID, "GLUCOSE_ALERT", fmt.Sprintf("Type: %s | Reading: %.0f mg/dL", alertType, glucoseReading))
}

func LogViewProfile(userID string) error {
	return LogAction(userID, "VIEW_PROFILE", "Accessed patient profile")
}

func LogViewLogs(userID string, patientID string, logType string) error {
	return LogAction(userID, "VIEW_LOGS", fmt.Sprintf("Patient: %s | Type: %s", patientID, logType))
}

func LogViewAlerts(userID string, patientID string) error {
	return LogAction(userID, "VIEW_ALERTS", fmt.Sprintf("Patient: %s", patientID))
}

func LogCaretakerBasalConfig(caretakerID string, patientID string, oldRate float64, newRate float64) error {
	return LogAction(caretakerID, "CARETAKER_BASAL_CONFIG", fmt.Sprintf("Patient: %s | Old: %.2f | New: %.2f units/hour (scheduled 24h)", patientID, oldRate, newRate))
}

func LogSystemEvent(eventType string, details string) error {
	return LogAction("SYSTEM", eventType, details)
}

func LogError(userID string, errorType string, errorMsg string) error {
	return LogAction(userID, fmt.Sprintf("ERROR_%s", errorType), errorMsg)
}

func LogSecurityEvent(userID string, eventDescription string) error {
	return LogAction(userID, "SECURITY_EVENT", eventDescription)
}

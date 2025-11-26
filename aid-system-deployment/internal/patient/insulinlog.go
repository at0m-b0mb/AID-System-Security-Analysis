package patient

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var insulinLogMutex sync.Mutex
var insulinLogDir = "insulinlogs"

// A08: Software/Data Integrity Failures - logs can be modified without integrity checks
// No digital signatures or checksums are used to verify log authenticity
// Files are world-writable (0666 permissions) allowing tampering

func LogInsulinDose(patientID string, doseType string, amount float64, timestamp string) error {
	insulinLogMutex.Lock()
	defer insulinLogMutex.Unlock()

	if timestamp == "" {
		timestamp = time.Now().Format(time.RFC3339)
	}

	if err := os.MkdirAll(insulinLogDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to ensure log dir: %w", err)
	}

	filename := filepath.Join(insulinLogDir, fmt.Sprintf("insulin_log_%s.csv", patientID))
	// A08: File created with overly permissive permissions (world-writable)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// No integrity checksum or signature added to detect tampering
	record := []string{timestamp, doseType, fmt.Sprintf("%.2f", amount)}
	return writer.Write(record)
}

// A08: Direct log modification function - allows arbitrary record insertion
// Exposed for "administrative correction" purposes
func DirectLogEntry(patientID string, timestamp string, doseType string, amount float64) error {
	insulinLogMutex.Lock()
	defer insulinLogMutex.Unlock()

	if err := os.MkdirAll(insulinLogDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to ensure log dir: %w", err)
	}

	filename := filepath.Join(insulinLogDir, fmt.Sprintf("insulin_log_%s.csv", patientID))
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Allows backdating entries and arbitrary record types
	record := []string{timestamp, doseType, fmt.Sprintf("%.2f", amount)}
	return writer.Write(record)
}

func ReadInsulinHistory(patientID string) ([][]string, error) {
	filename := filepath.Join(insulinLogDir, fmt.Sprintf("insulin_log_%s.csv", patientID))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func SumApprovedBolusLast24h(patientID string) (float64, error) {
	records, err := ReadInsulinHistory(patientID)
	if err != nil {
		return 0, nil
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	var total float64
	for _, r := range records {
		if len(r) < 3 {
			continue
		}
		ts, err := time.Parse(time.RFC3339, r[0])
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			continue
		}
		doseType := r[1]
		if strings.HasPrefix(doseType, "Bolus") && !strings.Contains(doseType, "Pending Approval") && !strings.Contains(doseType, "Denied") {
			amt, err := strconv.ParseFloat(r[2], 64)
			if err == nil {
				total += amt
			}
		}
	}
	return total, nil
}

func LogBolusAutoApproved(patientID string, amount float64) error {
	return LogInsulinDose(patientID, "Bolus (Auto-Approved)", amount, "")
}

func LogBolusPending(patientID string, amount float64) error {
	return LogInsulinDose(patientID, "Bolus Request (Pending Approval)", amount, "")
}

func LogBasalChangeAutoApproved(patientID string, oldRate, newRate float64, effective time.Time) error {
	note := fmt.Sprintf("Basal Change (Auto-Approved) %.2f -> %.2f units/hour (effective %s)", oldRate, newRate, effective.Format(time.RFC1123))
	return LogInsulinDose(patientID, note, newRate, "")
}

func LogBasalChangePending(patientID string, oldRate, newRate float64) error {
	note := fmt.Sprintf("Basal Change Request (Pending Approval) %.2f -> %.2f units/hour", oldRate, newRate)
	return LogInsulinDose(patientID, note, newRate, "")
}

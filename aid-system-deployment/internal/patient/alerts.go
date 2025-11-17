package patient

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Alert struct {
	Timestamp string
	Value     string
	Level     string
	Ack       bool
}

var alertMutex sync.Mutex
var lastAlertTimestamp string
var alertStopChan chan struct{}
var showPatientIDInAlerts bool

func InitAlerts() {
	alertStopChan = make(chan struct{})
}

func StopAlerts() {
	if alertStopChan != nil {
		close(alertStopChan)
	}
}

func GetAlertStopChan() chan struct{} {
	return alertStopChan
}

func SetAlertDisplayMode(showID bool) {
	showPatientIDInAlerts = showID
}

func AddAlertToCSV(patientID, timestamp, value, level string) error {
	alertMutex.Lock()
	defer alertMutex.Unlock()

	if err := os.MkdirAll("alerts", os.ModePerm); err != nil {
		return fmt.Errorf("failed to ensure alerts dir: %w", err)
	}

	filename := filepath.Join("alerts", fmt.Sprintf("alerts_log_%s.csv", patientID))
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{timestamp, value, level}
	return writer.Write(record)
}

func AlertHandler(patientID, timestamp, value, level string) {
	if timestamp == lastAlertTimestamp {
		return
	}
	lastAlertTimestamp = timestamp

	patient, err := GetPatientProfile(patientID)
	var patientName string
	if err != nil {
		patientName = patientID
	} else {
		patientName = patient.FullName
	}

	var msg string
	if showPatientIDInAlerts {
		msg = fmt.Sprintf("🚨 ALERT: Patient %s (%s) has %s blood sugar at %s: %s mg/dL", patientName, patientID, level, timestamp, value)
	} else {
		msg = fmt.Sprintf("🚨 ALERT: %s blood sugar at %s: %s mg/dL", level, timestamp, value)
	}
	fmt.Println(msg)

	err = AddAlertToCSV(patientID, timestamp, value, level)
	if err != nil {
		fmt.Println("Failed to log alert:", err)
	}
}

func ViewAlerts(patientID string) {
	filename := filepath.Join("alerts", fmt.Sprintf("alerts_log_%s.csv", patientID))
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Could not open alert log file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading alert log file:", err)
		return
	}

	fmt.Println("---- Alert History ----")
	for i, record := range records {
		if len(record) < 3 {
			continue
		}
		fmt.Printf("%d. [%s] %s at %s mg/dL\n", i+1, record[2], record[0], record[1])
	}
}

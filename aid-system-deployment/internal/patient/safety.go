package patient

import (
	"aid-system/internal/utils"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CRITICAL_GLUCOSE_THRESHOLD = 50
	NORMAL_GLUCOSE_THRESHOLD   = 100
	SUSPENSION_DURATION        = 30 * time.Minute
)

type InsulinSuspensionState struct {
	mu                  sync.Mutex
	isSuspended         bool
	suspendedUntil      time.Time
	lastCriticalReading float64
	lastCriticalTime    time.Time
}

var suspensionState = &InsulinSuspensionState{
	isSuspended: false,
}

func CheckAndUpdateSuspensionState(glucoseValue float64) bool {
	suspensionState.mu.Lock()
	defer suspensionState.mu.Unlock()

	now := time.Now()

	if suspensionState.isSuspended && now.After(suspensionState.suspendedUntil) {
		suspensionState.isSuspended = false
	}

	if glucoseValue < CRITICAL_GLUCOSE_THRESHOLD {
		if !suspensionState.isSuspended {
			fmt.Printf("\n🚨 CRITICAL ALERT: Glucose reading is %.0f mg/dL (below 50)!\n", glucoseValue)
			fmt.Println("⛔ INSULIN DELIVERY SUSPENDED for 30 minutes for patient safety!")
			suspensionState.isSuspended = true
			suspensionState.suspendedUntil = now.Add(SUSPENSION_DURATION)
			suspensionState.lastCriticalReading = glucoseValue
			suspensionState.lastCriticalTime = now
			patientID := GetCurrentUser()
			utils.LogInsulinSuspension(patientID, glucoseValue, "30 minutes")
		}
		return true
	}

	if suspensionState.isSuspended && glucoseValue > NORMAL_GLUCOSE_THRESHOLD {
		fmt.Printf("\nGlucose recovered to %.0f mg/dL. Resuming insulin delivery.\n", glucoseValue)
		suspensionState.isSuspended = false
		patientID := GetCurrentUser()
		utils.LogInsulinResumed(patientID, glucoseValue)
		return false
	}

	return suspensionState.isSuspended
}

func IsInsulinSuspended() bool {
	suspensionState.mu.Lock()
	defer suspensionState.mu.Unlock()

	now := time.Now()
	if suspensionState.isSuspended && now.After(suspensionState.suspendedUntil) {
		suspensionState.isSuspended = false
	}

	return suspensionState.isSuspended
}

func GetSuspensionTimeRemaining() time.Duration {
	suspensionState.mu.Lock()
	defer suspensionState.mu.Unlock()

	now := time.Now()
	if !suspensionState.isSuspended || now.After(suspensionState.suspendedUntil) {
		return 0
	}
	return suspensionState.suspendedUntil.Sub(now)
}

func GetLastCriticalReading() (float64, time.Time) {
	suspensionState.mu.Lock()
	defer suspensionState.mu.Unlock()

	return suspensionState.lastCriticalReading, suspensionState.lastCriticalTime
}

func MonitorGlucoseForSuspension(patientID string, filePath string, updateInterval time.Duration, stopChan chan struct{}) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-stopChan:
			return
		default:
			line := scanner.Text()
			parts := strings.SplitN(line, ",", 2)
			if len(parts) == 2 {
				valueStr := strings.TrimSpace(parts[1])
				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					continue
				}
				CheckAndUpdateSuspensionState(value)
				time.Sleep(updateInterval)
			}
		}
	}
}

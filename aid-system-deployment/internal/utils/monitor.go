package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type CGMReading struct {
	Timestamp string
	Value     string
}

func DisplayRecentCGMReadings(filePath string, updateInterval time.Duration, numReadings int) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Could not open file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0

	exitChan := make(chan struct{})

	go func() {
		buf := bufio.NewReader(os.Stdin)
		buf.ReadString('\n')
		close(exitChan)
	}()

	fmt.Println("Streaming CGM readings. Press Enter at any time to stop.")
	fmt.Println("(After stream finishes, press Enter to go back to menu.)")
	for scanner.Scan() {
		select {
		case <-exitChan:
			fmt.Println("\nExiting glucose readings view.")
			return
		default:
			if count >= numReadings {
				fmt.Println("\nAll readings displayed.")
				fmt.Println("Press Enter to go back to menu...")
				<-exitChan
				return
			}
			line := scanner.Text()
			parts := strings.SplitN(line, ",", 2)
			if len(parts) == 2 {
				fmt.Printf("Simulated CGM at %s: %s mg/dL\n", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				time.Sleep(updateInterval)
				count++
			}
		}
	}
}

func MonitorCGMFileAlerts(patientID string, filePath string, updateInterval time.Duration, alertFunc func(patientID, timestamp, value, level string), stopChan chan struct{}) {
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
				timestamp := strings.TrimSpace(parts[0])
				valueStr := strings.TrimSpace(parts[1])
				value, err := strconv.Atoi(valueStr)
				if err != nil {
					continue
				}
				if value < 70 {
					alertFunc(patientID, timestamp, valueStr, "LOW")
				}
				if value > 180 {
					alertFunc(patientID, timestamp, valueStr, "HIGH")
				}
				time.Sleep(updateInterval)
			}
		}
	}
}

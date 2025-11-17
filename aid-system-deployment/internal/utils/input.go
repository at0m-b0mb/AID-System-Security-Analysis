package utils

import (
	"bufio"
	"fmt"
	"strings"
)

func PromptYesNo(reader *bufio.Reader, promptText string) bool {
	for {
		fmt.Print(promptText)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		switch text {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter 'y' or 'n'.")
		}
	}
}

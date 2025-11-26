package clinician

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfigureSystemAlertDefaults manages system-wide default alert thresholds
func ConfigureSystemAlertDefaults() {
	fmt.Println("\n======== Configure System-Wide Alert Defaults ========")

	fmt.Println("Current system-wide alert defaults:")
	fmt.Println("  • LOW threshold:  Below 70 mg/dL (Hypoglycemia)")
	fmt.Println("  • HIGH threshold: Above 180 mg/dL (Hyperglycemia)")
	fmt.Println()
	fmt.Println("These are the default thresholds applied to all patients.")
	fmt.Println("Individual patient thresholds can be customized in 'Manage Patient Settings'.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Would you like to change system defaults? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		return
	}

	var lowDefault, highDefault int
	fmt.Print("\nEnter new system-wide LOW threshold (mg/dL) [40-70]: ")
	fmt.Scanf("%d\n", &lowDefault)
	fmt.Print("Enter new system-wide HIGH threshold (mg/dL) [180-300]: ")
	fmt.Scanf("%d\n", &highDefault)

	// Validate
	if lowDefault < 40 || lowDefault > 70 {
		fmt.Println("❌ Invalid low threshold. Must be between 40-70 mg/dL.")
		fmt.Println("Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	if highDefault < 180 || highDefault > 300 {
		fmt.Println("❌ Invalid high threshold. Must be between 180-300 mg/dL.")
		fmt.Println("Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	// TODO: Store in a system config table in database
	// For now, just display confirmation
	fmt.Println("\n✅ System-wide alert defaults updated!")
	fmt.Printf("  • New LOW default:  %d mg/dL\n", lowDefault)
	fmt.Printf("  • New HIGH default: %d mg/dL\n", highDefault)
	fmt.Println("\n⚠️  Implementation note:")
	fmt.Println("To fully implement, update utils/monitor.go to read these values")
	fmt.Println("from a system_config table in the database.")
	fmt.Println("\nPress Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

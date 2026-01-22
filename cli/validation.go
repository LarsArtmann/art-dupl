// Package cli provides command-line interface utilities.
package cli

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
)

// ExitIfBothSet prints error and exits if both flags are true.
// Used to simplify mutually exclusive flag checking.
func ExitIfBothSet(flag1, flag2 *bool, flag1Name, flag2Name string) int {
	if flag1 != nil && *flag1 && flag2 != nil && *flag2 {
		names := fmt.Sprintf("%s and %s", flag1Name, flag2Name)
		logger.Default.Error("you can have either output", "format", names)
		os.Exit(1)
		return 1
	}
	return 0
}

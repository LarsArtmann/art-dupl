package cmd

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// VersionInfo provides structured version information for CI tooling.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// NewVersionCommand creates the version subcommand.
func NewVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			short, _ := cmd.Flags().GetBool("short")
			if short {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version)
				return nil
			}

			jsonFormat, _ := cmd.Flags().GetBool("json")
			if jsonFormat {
				info := VersionInfo{
					Version:   Version,
					Commit:    Commit,
					Date:      Date,
					GoVersion: runtime.Version(),
					Compiler:  runtime.Compiler,
					Platform:  runtime.GOOS,
					Arch:      runtime.GOARCH,
				}
				data, err := json.Marshal(info, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
				if err != nil {
					return fmt.Errorf("marshal version info: %w", err)
				}

				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			PrintVersion()
			return nil
		},
	}

	versionCmd.Flags().Bool("json", false, "output version information as JSON")
	versionCmd.Flags().BoolP("short", "s", false, "print only the version string")

	return versionCmd
}

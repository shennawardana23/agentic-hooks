package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version, BuildTime, and GoVersion are set via -ldflags at build time
// (see the Makefile's LDFLAGS); they stay at these defaults under `go run`.
var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "agentic-hooks",
		Short: "Second Brain orchestration CLI",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the agentic-hooks version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "agentic-hooks %s (built %s, %s)\n", Version, BuildTime, GoVersion)
			return err
		},
	}

	root.AddCommand(versionCmd, newRunCmd(), newServeCmd())
	return root
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

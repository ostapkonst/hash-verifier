package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ostapkonst/hash-verifier/internal/gui"
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:           "hashverifier",
	Short:         "Cross-platform checksum generation and verification tool",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return gui.Run("")
		}

		return gui.Run(args[0])
	},
}

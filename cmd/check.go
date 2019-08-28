package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "count and report TS packets between delimiters",
	Long:  `Reports the number of packets betwen delimiters that are inserted after every write by file writers (set debug=true in writer when recording)`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of vw",
	Long:  `All software has versions. This is vw's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("vw video websocket transporter v0.1.0")
	},
}

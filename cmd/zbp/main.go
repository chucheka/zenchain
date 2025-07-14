package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var zbpCmd = &cobra.Command{
		Use:   "zbp",
		Short: "Zenblock Properties CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	zbpCmd.AddCommand(versionCmd)
	zbpCmd.AddCommand(balancesCmd())
	zbpCmd.AddCommand(txCmd())

	err := zbpCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}

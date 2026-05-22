package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a VPS host via interactive wizard",
	Long:  "Run the interactive wizard to generate keys, update ssh config, and optionally deploy a public key.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("add: not implemented yet")
	},
}

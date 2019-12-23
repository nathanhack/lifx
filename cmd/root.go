package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var ip string
var port int

var rootCmd = &cobra.Command{
	Use:   "lifx",
	Short: "Does things with LIFX bulbs",
	Long:  `Does things with LIFX bulbs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lightCmd)
	lightCmd.PersistentFlags().StringVar(&ip, "ip", "", "identify the IP if known")
	lightCmd.PersistentFlags().IntVar(&port, "port", 56700, "identify the Port if known")
}

var lightCmd = &cobra.Command{
	Use:   "light",
	Short: "Sends/receives light messages",
	Long:  `Light is a set of send/receive light messages.`,
}

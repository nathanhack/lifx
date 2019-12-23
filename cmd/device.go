package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deviceCmd)
	deviceCmd.PersistentFlags().StringVar(&ip, "ip", "", "identify the IP if known")
	deviceCmd.PersistentFlags().IntVar(&port, "port", -1, "identify the Port if known")
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Sends/receives device messages",
	Long:  `Device is a set of send/receive device messages.`,
}

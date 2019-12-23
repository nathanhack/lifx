package cmd

import (
	"context"
	"fmt"
	"github.com/nathanhack/lifx/cmd/gui"
	"github.com/nathanhack/lifx/core/server"
	"strings"

	"github.com/spf13/cobra"
)

var guiTargets []string

func init() {
	rootCmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui TARGET_HEXSTRING TARGET_HEXSTRING ...",
	Short: "Fullscreen GUI for regulating a group of lights",
	Long: `Fullscreen GUI for managing the list of TARGET_HEXSTRING.  
Ideally the lights should be apart of one groups.'`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("gui called", strings.Join(args, ","))
		ctx, cancel := context.WithCancel(context.Background())
		out, in, err := server.StartUp(ctx)
		if err != nil {
			return err
		}

		g := gui.GUI{
			Targets:  args,
			OutBound: out,
			Inbound:  in,
		}

		err = g.Run()
		cancel()
		return
	},
}

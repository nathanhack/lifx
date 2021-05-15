package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/nathanhack/lifx/cmd/internal"
	"github.com/nathanhack/lifx/core/broadcast"
	"github.com/nathanhack/lifx/core/header"
	"github.com/nathanhack/lifx/core/messages/light"
	"github.com/nathanhack/lifx/core/server"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var lightSetPowerOn bool
var lightSetPowerDuration uint32

func init() {
	lightCmd.AddCommand(lightSetPowerCmd)

	lightSetPowerCmd.Flags().BoolVar(&lightSetPowerOn, "on", false, "turns ON light (otherwise OFF)")
	lightSetPowerCmd.Flags().Uint32Var(&lightSetPowerDuration, "duration", 0, "transition time in milliseconds")
}

var lightSetPowerCmd = &cobra.Command{
	Use:   "setpower TARGET_HEXSTR [TIMEOUT_MILLISECONDS]",
	Short: "Sets the power ON or OFF (standby) for a particular LIFX light",
	Long: `Sets the power ON or OFF (standby) for a particular LIFX light identified by TARGET_HEXSTR. 

Note if the IP and port are known include those tags then the broadcast step can be skipped.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := light.SetPower{Duration: lightSetPowerDuration}
		message.SetLevel(lightSetPowerOn)

		timeout := defaultTimeout
		if len(args) > 1 {
			tmp, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			timeout = time.Duration(tmp) * time.Millisecond
			fmt.Printf("Timeout found, using %v\n", timeout)
		}
		ctx := context.Background()
		out, in, err := server.StartUp(ctx)
		if err != nil {
			return err
		}
		var targetBroadcast *broadcast.BroadcastResult
		if ip == "" || port <= 0 {
			bctx, _ := context.WithTimeout(ctx, timeout)
			targetBroadcasts, err := sendBroadcast(bctx, out, in, []string{args[0]}, false)
			if err != nil {
				return err
			}
			if len(targetBroadcasts) == 0 {
				fmt.Println("could not find target device")
				return nil
			}
			targetBroadcast = targetBroadcasts[args[0]]
		} else {
			address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
			if err != nil {
				return err
			}
			hexBytes, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			targetBroadcast = &broadcast.BroadcastResult{
				Target: hexBytes,
				IP:     address.IP,
				Port:   address.Port,
			}
		}

		//now lets get the powerlevel
		pctx, _ := context.WithTimeout(ctx, timeout)
		err = sendLightSetPower(pctx, out, targetBroadcast, &message)
		if err != nil {
			return err
		}

		return nil
	},
}

func sendLightSetPower(ctx context.Context, out chan *server.OutBoundPayload, targetBroadcast *broadcast.BroadcastResult, message *light.SetPower) (err error) {
	head := header.New(internal.GetNextSequence())
	head.SetTarget(targetBroadcast.Target)
	message.RequiredHeader(head, true)
	buffer := bytes.NewBuffer([]byte{})
	err = binary.Write(buffer, binary.LittleEndian, head)
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.LittleEndian, message)
	if err != nil {
		return err
	}

	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", targetBroadcast.IP, targetBroadcast.Port))
	if err != nil {
		return err
	}

	tmp := "OFF"
	if message.GetLevel() {
		tmp = "ON"
	}

	fmt.Printf("Setting power to %02x at %v:%v to %v\n", targetBroadcast.Target, targetBroadcast.IP, targetBroadcast.Port, tmp)
	localctx, done := context.WithCancel(ctx)
	out <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: address,
		Done:    done,
	}
	// now we wait for receipt that the message has been sent
	<-localctx.Done()
	return nil
}

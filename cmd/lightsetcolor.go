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

var (
	lightSetColorHue      int
	lightSetColorSat      int
	lightSetColorBright   int
	lightSetColorKelvin   int
	lightSetColorDuration uint32
)

func init() {
	lightCmd.AddCommand(lightSetColorCmd)

	lightSetColorCmd.Flags().IntVar(&lightSetColorHue, "hue", -1, "sets hue value [0,360]")
	lightSetColorCmd.Flags().IntVarP(&lightSetColorSat, "saturation", "s", -1, "sets saturation value [0,100]")
	lightSetColorCmd.Flags().IntVarP(&lightSetColorBright, "brightness", "b", -1, "set brightness value [0,100]")
	lightSetColorCmd.Flags().IntVarP(&lightSetColorKelvin, "kelvin", "k", -1, "set kelvin value [2500,9000]")
	lightSetColorCmd.Flags().Uint32VarP(&lightSetColorDuration, "duration", "d", 0, "time in milliseconds for transition")
}

func intmin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intmax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minMaxInt(x, max, min int) int {
	return intmax(intmin(max, x), min)
}

var lightSetColorCmd = &cobra.Command{
	Use:   "setcolor TARGET_HEXSTR  [TIMEOUT_MILLISECONDS]",
	Short: "Sets the color for a particular LIFX light",
	Long: `Sets the color for the LIFX light identified by TARGET_HEXSTR.

Note if the IP and port are known include those tags then the broadcast step can be skipped.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		timeout := defaultTimeout
		if len(args) > 2 {
			tmp, err := strconv.Atoi(args[2])
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

		message := light.SetColor{}
		message.Color.Hue = uint16(minMaxInt(int(float32(lightSetColorHue)/360*0xffff), 0xffff, 0))
		message.Color.Saturation = uint16(minMaxInt(int(float32(lightSetColorSat)/100*0xffff), 0xffff, 0))
		message.Color.Brightness = uint16(minMaxInt(int(float32(lightSetColorBright)/100*0xffff), 0xffff, 0))
		message.Color.Kelvin = uint16(minMaxInt(lightSetColorKelvin, 9000, 2400))
		message.Duration = lightSetColorDuration

		// now we need to replace any negative values with the current state from the light
		if lightSetColorHue < 0 || lightSetColorSat < 0 || lightSetColorBright < 0 || lightSetColorKelvin < 0 {
			bctx, _ := context.WithTimeout(ctx, timeout)
			oldstate, err := sendLightGet(bctx, out, in, targetBroadcast)
			if err != nil {
				return err
			}
			if lightSetColorHue < 0 {
				message.Color.Hue = oldstate.Color.Hue
			}
			if lightSetColorSat < 0 {
				message.Color.Saturation = oldstate.Color.Saturation
			}
			if lightSetColorBright < 0 {
				message.Color.Brightness = oldstate.Color.Brightness
			}
			if lightSetColorKelvin < 0 {
				message.Color.Kelvin = oldstate.Color.Kelvin
			}
		}

		//now lets get the powerlevel
		pctx, _ := context.WithTimeout(ctx, timeout)
		err = sendLightSetColor(pctx, out, targetBroadcast, &message)
		if err != nil {
			return err
		}

		return nil
	},
}

func sendLightSetColor(ctx context.Context, out chan *server.OutBoundPayload, targetBroadcast *broadcast.BroadcastResult, message *light.SetColor) error {
	head := header.New(internal.GetNextSequence())
	head.SetTarget(targetBroadcast.Target)
	message.RequiredHeader(head, true)
	buffer := bytes.NewBuffer([]byte{})
	err := binary.Write(buffer, binary.LittleEndian, head)
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

	fmt.Printf("Setting color to %02x at %v:%v to %v\n", targetBroadcast.Target, targetBroadcast.IP, targetBroadcast.Port, message.Color)
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

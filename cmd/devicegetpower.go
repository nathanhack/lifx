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
	"github.com/nathanhack/lifx/core/messages/device"
	"github.com/nathanhack/lifx/core/server"
	"github.com/spf13/cobra"
	"net"
	"reflect"
	"strconv"
	"time"
)

func init() {
	deviceCmd.AddCommand(deviceGetPowerCmd)
}

var deviceGetPowerCmd = &cobra.Command{
	Use:   "getpower TARGET_HEXSTR [TIMEOUT_MILLISECONDS]",
	Short: "Retrieves the power status info for a particular LIFX device",
	Long: `Retrieves the power status info for the LIFX device identified by TARGET_HEXSTR. The status includes power level.

Note if the IP and port are known include those tags then the broadcast step can be skipped.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		pstate, err := sendDeviceGetPower(pctx, out, in, targetBroadcast)
		if err != nil {
			return err
		}

		fmt.Println(pstate.Level)

		return nil
	},
}

func sendDeviceGetPower(ctx context.Context, out chan *server.OutBoundPayload, in chan *server.InboundPayload, targetBroadcast *broadcast.BroadcastResult) (state *device.StatePower, err error) {
	head := header.New(internal.GetNextSequence())
	head.SetTarget(targetBroadcast.Target)
	message := device.GetPower{}
	message.RequiredHeader(head)
	buffer := bytes.NewBuffer([]byte{})
	err = binary.Write(buffer, binary.LittleEndian, head)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, &message)
	if err != nil {
		return nil, err
	}

	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", targetBroadcast.IP, targetBroadcast.Port))
	if err != nil {
		return nil, err
	}

	fmt.Printf("Sending GetPower Request to %v:%v\n", targetBroadcast.IP, targetBroadcast.Port)
	out <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: address,
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Timeout")
		case payload := <-in:
			h, err := header.Decode(payload.Data)
			if err != nil {
				continue
			}
			if !h.Validate(true) {
				continue
			}

			if reflect.DeepEqual(h.Target(), targetBroadcast.Target) &&
				h.Type() == device.StatePowerType {
				var s device.StatePower
				err = device.DecodeFromHeader(*h, &s)
				if err != nil {
					continue
				}
				state = &s

				return state, nil
			}
		}
	}
}

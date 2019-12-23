package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/nathanhack/lifx/core/broadcast"
	"github.com/nathanhack/lifx/core/header"
	"github.com/nathanhack/lifx/core/messages/device"
	"github.com/nathanhack/lifx/core/server"
	"github.com/spf13/cobra"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
)

var broadcastShowLabel bool

func init() {
	rootCmd.AddCommand(broadcastCmd)
	broadcastCmd.Flags().BoolVar(&broadcastShowLabel, "label", false, "show label")

}

var broadcastCmd = &cobra.Command{
	Use:   "broadcast [TIMEOUT_MILLISECONDS]",
	Short: "Sends a broadcast request",
	Long:  `Sends a broadcast request and prints the broadcast results. If TIMEOUT_MILLISECONDS not specified then 5000 milliseconds (5s) is used.`,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout := defaultTimeout
		if len(args) > 0 {
			tmp, err := strconv.Atoi(args[0])
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

		bctx, _ := context.WithTimeout(ctx, timeout)
		_, err = sendBroadcast(bctx, out, in, []string{}, broadcastShowLabel)
		if err != nil {
			return err
		}

		return nil
	},
}

func sendBroadcast(ctx context.Context, out chan *server.OutBoundPayload, in chan *server.InboundPayload, filterByHexString []string, requestLabels bool) (targetBroadcasts map[string]*broadcast.BroadcastResult, err error) {
	//we'll make a map to hold the filterByHexString's byte versions
	hexStringsFilter := make(map[string][]byte)
	for _, s := range filterByHexString {
		hexBytes, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("error while decoding %v: %v", s, err)
		}
		hexStringsFilter[strings.ToLower(s)] = hexBytes
	}

	head := header.New()
	message := device.GetService{}
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
	broadcastAddress, err := net.ResolveUDPAddr("udp", "255.255.255.255:56700")
	if err != nil {
		return nil, err
	}

	fmt.Println("sending broadcast to determine address for device(s)")
	out <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: broadcastAddress,
	}

	//if we have a hexString filters then we return as soon as we got them all
	targetBroadcasts = make(map[string]*broadcast.BroadcastResult)

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-in:
			h, err := header.Decode(payload.Data)
			if err != nil {
				continue
			}
			if !h.Validate(true) {
				continue
			}

			hexStr := hex.EncodeToString(h.Target())
			targetBroadcast := &broadcast.BroadcastResult{
				Target: h.Target(),
				IP:     payload.Conn.IP,
				Port:   payload.Conn.Port,
			}

			switch h.Type() {
			case device.StateServiceType:
				if len(hexStringsFilter) > 0 {
					if _, has := hexStringsFilter[hexStr]; has {
						targetBroadcasts[hexStr] = targetBroadcast
						fmt.Printf("Found target %x at %v:%v\n", targetBroadcast.Target, payload.Conn.IP, payload.Conn.Port)
					}
				} else {
					if !requestLabels {
						if _, has := targetBroadcasts[hexStr]; !has {
							targetBroadcasts[hexStr] = targetBroadcast
							fmt.Printf("Found LIFX %x at %v:%v\n", targetBroadcast.Target, payload.Conn.IP, payload.Conn.Port)
						}
					} else {
						sendDeviceGetLabel(out, targetBroadcast.Target, payload.Conn.IP, payload.Conn.Port)
					}
				}

				if len(filterByHexString) > 0 && len(filterByHexString) == len(targetBroadcasts) {
					return targetBroadcasts, nil
				}

			case device.StateLabelType:
				var s device.StateLabel
				err := device.DecodeFromHeader(*h, &s)
				if err != nil {
					continue
				}
				if _, has := targetBroadcasts[hexStr]; !has {
					targetBroadcasts[hexStr] = targetBroadcast
					fmt.Printf("Found LIFX %x (%v) at %v:%v\n", targetBroadcast.Target, s.GetLabel(), payload.Conn.IP, payload.Conn.Port)
				}
			}
		}
	}
}

func sendDeviceGetLabel(out chan *server.OutBoundPayload, target []byte, ip net.IP, port int) error {
	head := header.New()
	head.SetTarget(target)
	message := device.GetLabel{}
	message.RequiredHeader(head)
	buffer := bytes.NewBuffer([]byte{})
	err := binary.Write(buffer, binary.LittleEndian, head)
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.LittleEndian, &message)
	if err != nil {
		return err
	}

	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		return err
	}

	out <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: address,
	}
	return nil
}

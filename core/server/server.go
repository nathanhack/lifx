package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type InboundPayload struct {
	Data []byte
	Conn *net.UDPAddr
}

type OutBoundPayload struct {
	Data    []byte
	Address *net.UDPAddr
	Done    context.CancelFunc
}

func StartUp(ctx context.Context) (outBound chan *OutBoundPayload, inbound chan *InboundPayload, err error) {
	outBound = make(chan *OutBoundPayload, 10)
	inbound = make(chan *InboundPayload, 100)

	localAddress, _ := net.ResolveUDPAddr("udp", ":56700")
	connection, err := net.ListenUDP("udp", localAddress)
	if err != nil {
		logrus.Errorf("%v", err)
		return nil, nil, err
	}

	go func() {
		defer connection.Close()
		logrus.Debugf("starting server outbound")
		for {
			select {
			case <-ctx.Done():
				return
			case payload := <-outBound:
				connection.WriteTo(payload.Data, payload.Address)
				if payload.Done != nil {
					payload.Done()
				}
			}
		}
	}()

	go func() {
		defer connection.Close()
		logrus.Debugf("starting server inbound")
		for {
			select {
			case <-ctx.Done():
				return
			default:

			}
			inputBytes := make([]byte, 2048)
			connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			length, conn, err := connection.ReadFromUDP(inputBytes)
			if err != nil {
				continue
			}
			inbound <- &InboundPayload{
				Data: inputBytes[:length],
				Conn: conn,
			}
		}
	}()

	return outBound, inbound, nil
}

package gui

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/nathanhack/lifx/cmd/internal"
	"github.com/nathanhack/lifx/core/header"
	"github.com/nathanhack/lifx/core/messages/device"
	"github.com/nathanhack/lifx/core/messages/light"
	"github.com/nathanhack/lifx/core/server"
	"github.com/sirupsen/logrus"
	"image/color"
	"log"
	"math"
	"net"
	"time"
)

var (
	normalTermination = fmt.Errorf("normal termination")
	emptyImage, _     = ebiten.NewImage(1920, 1080, ebiten.FilterDefault)
	options           = &ebiten.DrawTrianglesOptions{}
)

func init() {
	emptyImage.Fill(color.White)
	var err error
	tt, err = truetype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
}

type targetString string

type GUI struct {
	Targets          []string
	OutBound         chan *server.OutBoundPayload
	Inbound          chan *server.InboundPayload
	Width            int
	Height           int
	lights           map[targetString]*guiLight
	bigLight         guiLight
	screenSaver      bool
	screenSaverLight guiLight
	lastInteraction  time.Time
}

func (gui *GUI) Run() error {
	ebiten.SetFullscreen(true)

	gui.Width, gui.Height = ebiten.ScreenSizeInFullscreen()
	// On mobiles, ebiten.MonitorSize is not available so far.
	// Use arbitrary values.
	if gui.Width == 0 || gui.Height == 0 {
		gui.Width = 300
		gui.Height = 450
	}

	go func() {
	mainLoop:
		for {
			select {
			case <-time.After(5 * time.Second):
				//every five seconds we check each light
				//first we'll make sure we have updated addresses for all the lights
				for _, l := range gui.lights {
					if l.address == nil || time.Since(l.lastAddressUpdate) > 24*time.Hour {
						// we'll send out another broadcast
						sendBroadcast(gui.OutBound)
						continue mainLoop
					}
				}

				// and see if when was the last time we saw data
				for _, l := range gui.lights {
					if time.Since(l.lastSeen) > 5*time.Minute {
						// we'll send out another broadcast
						sendDeviceGetPower(gui.OutBound, l.address)
						sendLightGet(gui.OutBound, l.address)
						continue mainLoop
					}
				}
			case payload := <-gui.Inbound:
				h, err := header.Decode(payload.Data)
				if err != nil || !h.Validate(true) {
					continue
				}
				if _, has := gui.lights[targetString(h.TargetHex())]; !has {
					continue
				}

				//now based on the type we'll do something with it
				switch h.Type() {
				case device.StateServiceType:
					//this response happens after a broadcast
					//we update the last seen time on the light
					gui.lights[targetString(h.TargetHex())].lastAddressUpdate = time.Now()
					gui.lights[targetString(h.TargetHex())].address = payload.Conn

					// so we'll send out a power level request
					if err := sendDeviceGetPower(gui.OutBound, payload.Conn); err != nil {
						logrus.Error(err)
						gui.Error()
					}

					//we also want to know how much it's on
					if err := sendLightGet(gui.OutBound, payload.Conn); err != nil {
						logrus.Error(err)
						gui.Error()
					}

				case device.StatePowerType:
					//when we get a power it we update the light's state
					var state device.StatePower
					if err := device.DecodeFromHeader(*h, &state); err != nil {
						logrus.Error(err)
						gui.Error()

						if err := sendDeviceGetPower(gui.OutBound, payload.Conn); err != nil {
							logrus.Error(err)
						}
						return
					}
					gui.lights[targetString(string(h.TargetHex()))].lastSeen = time.Now()
					gui.lights[targetString(h.TargetHex())].SetOn(state.GetLevel())
					gui.updateBigLight()
					gui.updateScreenSaverLight()

				case light.StateType:
					var state light.State
					if err := light.DecodeFromHeader(*h, &state); err != nil {
						logrus.Error(err)
						gui.Error()

						if err := sendDeviceGetPower(gui.OutBound, payload.Conn); err != nil {
							logrus.Error(err)
						}
						return
					}
					gui.lights[targetString(string(h.TargetHex()))].lastSeen = time.Now()
					gui.lights[targetString(h.TargetHex())].label = state.GetLabel()
					gui.lights[targetString(h.TargetHex())].SetOn(state.GetPower())
					gui.updateBigLight()
					gui.updateScreenSaverLight()
				}

			}
		}
	}()

	s := ebiten.DeviceScaleFactor()
	if err := ebiten.Run(gui.update, int(float64(gui.Width)*s), int(float64(gui.Height)*s), 1/s, "LIFX GUI Manager"); err != nil && err != normalTermination {
		return err
	}
	return nil
}

func (gui *GUI) update(screen *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return normalTermination
	}
	if len(gui.lights) == 0 {
		gui.lights = make(map[targetString]*guiLight)
		num := float64(len(gui.Targets))
		for i, target := range gui.Targets {
			gui.lights[targetString(target)] = &guiLight{
				x:      float32(gui.Width)/2 + float32((500+100)/2*math.Cos(2*math.Pi/num*float64(i))),
				y:      float32(gui.Height)/2 + float32((500+100)/2*math.Sin(2*math.Pi/num*float64(i))),
				size:   200,
				simple: true,
				on:     false,
			}
		}
		gui.bigLight = guiLight{
			x:      float32(gui.Width) / 2,
			y:      float32(gui.Height) / 2,
			size:   500,
			num:    len(gui.Targets),
			simple: false,
			on:     false,
		}

		gui.screenSaverLight = guiLight{
			x:      float32(gui.Width) / 2,
			y:      float32(gui.Height) / 2,
			size:   200,
			num:    len(gui.Targets),
			simple: false,
			on:     false,
		}
		//well we don't really know the state so
		// we send out a broadcast to get the state

		if err := sendBroadcast(gui.OutBound); err != nil {
			logrus.Error(err)
			gui.Error()
		}
	}

	if time.Since(gui.lastInteraction) > 15*time.Second {
		return gui.updateScreenSaver(screen)
	}
	return gui.updateNormalScreen(screen)
}

func (gui *GUI) updateScreenSaver(screen *ebiten.Image) error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		gui.lastInteraction = time.Now()
		return nil
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	//if not half way down move it
	if gui.screenSaverLight.y != float32(gui.Height/2) {
		gui.screenSaverLight.MoveBy(0, float32(gui.Height/2)-gui.screenSaverLight.y)
	}
	//we do side to side movements 60 seconds per cycle

	onecycle := (60 * time.Second).Nanoseconds()
	now := time.Now().UnixNano()
	x1 := float32(math.Cos(2*math.Pi/float64(onecycle)*float64(now%onecycle)))*(float32(gui.Width)/2-gui.screenSaverLight.size/2) + float32(gui.Width)/2
	x0 := gui.screenSaverLight.x
	gui.screenSaverLight.MoveBy(x1-x0, 0)
	gui.screenSaverLight.Update(screen)

	return nil
}

func (gui *GUI) updateNormalScreen(screen *ebiten.Image) error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		for _, l := range gui.lights {
			if l.In(float32(mx), float32(my)) {
				if time.Since(l.lastChange) > FadingTime*2 {
					l.lastChange = time.Now()
					sendDeviceSetPower(gui.OutBound, l.address, !l.on)
					go func() {
						time.Sleep(FadingTime*2 + 100*time.Millisecond)
						sendLightGet(gui.OutBound, l.address)
					}()
					l.SetOn(!l.on)
					gui.updateBigLight()
					gui.lastInteraction = time.Now()
				}
				return nil
			}
		}
		//at this point we should check if we've clicked the biglight
		if gui.bigLight.In(float32(mx), float32(my)) {
			if time.Since(gui.bigLight.lastChange) > FadingTime*2 {
				gui.bigLight.lastChange = time.Now()
				for _, l := range gui.lights {
					sendDeviceSetPower(gui.OutBound, l.address, !gui.bigLight.on)
				}
				go func() {
					time.Sleep(FadingTime*2 + 100*time.Millisecond)
					for _, l := range gui.lights {
						sendLightGet(gui.OutBound, l.address)
					}
				}()
				gui.bigLight.SetOn(!gui.bigLight.on)
				gui.lastInteraction = time.Now()
			}
			return nil
		}

	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	gui.bigLight.Update(screen)
	for _, l := range gui.lights {
		l.Update(screen)
	}
	return nil
}

func (gui *GUI) Error() {
	go func() {
		gui.bigLight.SetErr(true)
		time.Sleep(3 * time.Second)
		gui.bigLight.SetErr(false)
	}()
}

func (gui *GUI) updateBigLight() {
	on := false
	for _, l := range gui.lights {
		on = on || l.on
	}

	gui.bigLight.SetOn(on)
}

func (gui *GUI) updateScreenSaverLight() {
	on := false
	for _, l := range gui.lights {
		on = on || l.on
	}
	gui.screenSaverLight.SetOn(on)
}

func sendBroadcast(outbound chan *server.OutBoundPayload) error {
	head := header.New(internal.GetNextSequence())
	message := device.GetService{}
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

	broadcastAddress, err := net.ResolveUDPAddr("udp", "255.255.255.255:56700")
	if err != nil {
		return err
	}

	outbound <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: broadcastAddress,
	}
	return nil
}

func sendDeviceGetPower(outbound chan *server.OutBoundPayload, address *net.UDPAddr) error {
	head := header.New(internal.GetNextSequence())
	head.SetAcknowledgementRequired(true)
	message := device.GetPower{}
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

	outbound <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: address,
	}
	return nil
}

func sendLightGet(outbound chan *server.OutBoundPayload, address *net.UDPAddr) error {
	head := header.New(internal.GetNextSequence())
	message := light.Get{}
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

	outbound <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: address,
	}
	return nil
}

func sendDeviceSetPower(outbound chan *server.OutBoundPayload, address *net.UDPAddr, on bool) error {
	head := header.New(internal.GetNextSequence())
	message := device.SetPower{}
	message.RequiredHeader(head, false)
	message.SetLevel(on)
	buffer := bytes.NewBuffer([]byte{})

	err := binary.Write(buffer, binary.LittleEndian, head)
	if err != nil {
		return err
	}

	err = binary.Write(buffer, binary.LittleEndian, &message)
	if err != nil {
		return err
	}

	outbound <- &server.OutBoundPayload{
		Data:    buffer.Bytes(),
		Address: address,
	}
	return nil
}

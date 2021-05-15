# lifx
lifx is a GUI, commandline tool, and library for Lifx devices.

### Deps
Golang 1.13+

### GUI with Screensaver

<span>
<img src="https://user-images.githubusercontent.com/9204400/71376680-44528b00-2590-11ea-9ef7-e9da252b6f60.png" width="400" alt="GUI" title="GUI">

<img src="https://user-images.githubusercontent.com/9204400/71376725-6d731b80-2590-11ea-9b00-0e79ff983d9e.png" width="400" alt="Screensaver" title="Screensaver">
</span>
    

Running with `gui` at the commandline will start a fullscreen opengl program.  It's  purpose is to manage a list of TARGET lifx devices passed in via commandline args.

Example: `go run lifx.go gui d1234567891100 d1234567891200 d1234567891300`.

### Commandline

At the commandline there exists several commands based on the [LIFX API](https://lan.developer.lifx.com/docs).

Examples:
```
gu run lifx.go broadcast
go run lifx.go broadcast --label
go run lifx.go light get d1234567891100
go run lifx.go light getpower d1234567891100 
go run lifx.go light setpower d1234567891100 --duration 5000 --on
go run lifx.go light setcolor d1234567891100 --ip 192.168.0.100 --port 56700 --saturation 39 --hue 82
```   

Note: In order to determine the TARGETs use `broadcast` first to get a list.


### Library
If the GUI and commandline features aren't useful it can also be used as a library.  The more agnostic pieces can be found under the `core` directory.

##### Notes
The library was built to support the GUI, so there are gaps in the implemented messages it can send. If other messages are needed put in an issue or put in a PR ;-).

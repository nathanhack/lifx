# lifx
lifx is a GUI, commandline tool, and library for Lifx devices.

### Tasked GUI
Either build lifx or run with `gui` at the commandline will start a fullscreen opengl program.  It's  purpose is to manage a list of TARGET lifx devices passed in via commandline args.

Example: `go run lifx.go gui d1234567891100 d1234567891200`.

### Commandline

At the commandline there exists several commands based on the [LIFX API](https://lan.developer.lifx.com/docs).

Examples:
```
gu run lifx.go broadcast
go run lifx.go broadcast --label
go run lifx.go light get d1234567891100
go run lifx.go light getpower d1234567891100 
go run lifx.go light setpower d1234567891100 --duration 5000 --on
go run lifx.go light setcolor d1234567891100 --ip 10.0.0.31 --port 56700 --saturation 39 --hue 82
```   

Note: In order to determine the TARGETs use `broadcast` first to get a list.


### Library
If the GUI and commandline features aren't useful it can also be used as a library.  The more agnostic pieces can be found under the `core` directory.

##### Notes
The library was built to support the GUI, so there are gaps in the implemented messages it can send. If other messages are needed put in an issue or put in a PR ;-).
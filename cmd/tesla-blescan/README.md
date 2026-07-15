# Tesla BLE Scan Utility

The `tesla-blescan` application provides a command-line interface to find Tesla
vehicles in BLE range with some status information and RSSI.

This application does not run on Windows due to limitations in the available
Golang BLE packages.

## Building

Run `go get` to install Golang dependencies, and `go build` to compile.

You may also run `go install` to place `tesla-blescan` in your GOBIN directory.

## Key management

This application only uses the 'body-controller-state' command that does not 
require any authentication or access to a private key.

## Scanning for Tesla vehicles nearby

To scan for Tesla vehicles in BLE range simply run:

```
tesla-blescan 
```

Run `tesla-blescan -h` to see a full list of supported commands.

# Tesla BLE Scan Utility

The `tesla-blescan` application provides a command-line interface to find Tesla
vehicles in BLE range with some status information and RSSI.

This application does not run on Windows due to limitations in the available
Golang BLE packages.

## Building

Run `go get` to install Golang dependencies, and `go build` to compile.

You may also run `go install` to place `tesla-blescan` in your GOBIN directory.

## Key management

This application only supports the 'body-controller-state' and 'list-keys' commands, as they operate without requiring authentication or private key access.

## Scanning for Tesla vehicles nearby

To scan for Tesla vehicles in BLE range simply run:

```
tesla-blescan body-controller-state
{"scanResults":[{"localName":"S907xxxxxxxxxxxxbC","rssi":-79,"state":{"vehicleLockState":1, "vehicleSleepStatus":2, "userPresence":1}}
```

'localName' is a local name used in BLE communication and explained in the tesla-control utility.

The following examples are using jq - json query, a tool to filter, parse, format and transform json queries.

```
tesla-blescan body-controller-state enums | jq
{
  "scanResults": [
    {
      "localName": "S907xxxxxxxxxxxxbC",
      "rssi": -86,
      "state": {
        "vehicleLockState": "VEHICLELOCKSTATE_LOCKED",
        "vehicleSleepStatus": "VEHICLE_SLEEP_STATUS_ASLEEP",
        "userPresence": "VEHICLE_USER_PRESENCE_NOT_PRESENT"
      }
    }
  ]
}
```

Run `tesla-blescan -h` to see a full list of supported commands and options.

Here's an example of the 'list-keys' command:

```
 tesla-blescan list-keys | jq
{
  "scanResults": [
    {
      "localName": "S907xxxxxxxxxxxxbC",
      "rssi": -80,
      "keylist": [
        {
          "publicKey": "04dbxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxada4",
          "role": "ROLE_SERVICE",
          "formFactor": "KEY_FORM_FACTOR_UNKNOWN"
        },
        {
          "publicKey": "04a2xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxf750",
          "role": "ROLE_OWNER",
          "formFactor": "KEY_FORM_FACTOR_NFC_CARD"
        },
        {
          "publicKey": "040bxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx89eb",
          "role": "ROLE_OWNER",
          "formFactor": "KEY_FORM_FACTOR_NFC_CARD"
        },
        {
          "publicKey": "0438xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxabbe",
          "role": "ROLE_OWNER",
          "formFactor": "KEY_FORM_FACTOR_CLOUD_KEY"
        },
        {
          "publicKey": "04adxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxf02b",
          "role": "ROLE_OWNER",
          "formFactor": "KEY_FORM_FACTOR_CLOUD_KEY"
        },
        {
          "publicKey": "040exxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxc5e0",
          "role": "ROLE_OWNER",
          "formFactor": "KEY_FORM_FACTOR_IOS_DEVICE"
        },
        {
          "publicKey": "04d9xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx504b",
          "role": "ROLE_DRIVER",
          "formFactor": "KEY_FORM_FACTOR_ANDROID_DEVICE"
        },
        {
          "publicKey": "047dxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx0652",
          "role": "ROLE_OWNER",
          "formFactor": "KEY_FORM_FACTOR_IOS_DEVICE"
        },
        {
          "publicKey": "0457xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx39f52",
          "role": "ROLE_DRIVER",
          "formFactor": "KEY_FORM_FACTOR_ANDROID_DEVICE"
        }
      ]
    }
  ]
}
```

/*
tesla-blescan provides a command-line interface to search for Tesla vehicles in Bluetooth Low Entergy (BLE) range.
It shows the BLE name, address and RSSI for all vehicles found. All Tesla vehicles send BLE beacons with the
following format: SxxxxxxxxxxxxxxxxC where xxxxxxxxxxxxxxx is a 16-digit hexadecimal value that is
calculated from the first 8 bytes of the SHA1 checksum of the VIN.
A reverse lookup from the BLE name to a VIN would be possible with a large rainbow table, but not implemented so far.

The purpose of this tool is to find any vehicles in BLE range, show their RSSI and all information that is available by
the 'body-controller-state' command. See 'tesla-control' utility for details.
*/
package main

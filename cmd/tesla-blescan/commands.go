package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/teslamotors/vehicle-command/internal/log"
	"github.com/teslamotors/vehicle-command/pkg/account"
	"github.com/teslamotors/vehicle-command/pkg/cli"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
	"github.com/teslamotors/vehicle-command/pkg/protocol/protobuf/vcsec"
	"github.com/teslamotors/vehicle-command/pkg/vehicle"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ErrCommandLineArgs = errors.New("invalid command line arguments")
	ErrInvalidTime     = errors.New("invalid time")
	dayNamesBitMask    = map[string]int32{
		"SUN":       1,
		"SUNDAY":    1,
		"MON":       2,
		"MONDAY":    2,
		"TUES":      4,
		"TUESDAY":   4,
		"WED":       8,
		"WEDNESDAY": 8,
		"THURS":     16,
		"THURSDAY":  16,
		"FRI":       32,
		"FRIDAY":    32,
		"SAT":       64,
		"SATURDAY":  64,
		"ALL":       127,
		"WEEKDAYS":  62,
	}
)

type Argument struct {
	name string
	help string
}

type Handler func(ctx context.Context, acct *account.Account, car *vehicle.Vehicle, args map[string]string) error

type Command struct {
	help             string
	requiresAuth     bool // True if command requires client-to-vehicle authentication (private key)
	requiresFleetAPI bool // True if command requires client-to-server authentication (OAuth token)
	args             []Argument
	optional         []Argument
	handler          Handler
	domain           protocol.Domain
}

// configureAndVerifyFlags verifies that c contains all the information required to execute a command.
func configureFlags(c *cli.Config, commandName string, forceBLE bool) error {
	info, ok := commands[commandName]
	if !ok {
		return ErrUnknownCommand
	}
	c.Flags = cli.FlagBLE
	if info.domain != protocol.DomainNone {
		c.Domains = cli.DomainList{info.domain}
	}
	bleWake := forceBLE && commandName == "wake"
	if bleWake || info.requiresAuth {
		// Wake commands are special. When sending a wake command over the Internet, infotainment
		// cannot authenticate the command because it's asleep. When sending the command over BLE,
		// VCSEC _does_ authenticate the command before poking infotainment.
		c.Flags |= cli.FlagPrivateKey | cli.FlagVIN
	}
	if bleWake {
		// Normally, clients send out two handshake messages in parallel in order to reduce latency.
		// One handshake with VCSEC, one handshake with infotainment. However, if we're sending a
		// BLE wake command, then infotainment is (presumably) asleep, and so we should only try to
		// handshake with VCSEC.
		c.Domains = cli.DomainList{protocol.DomainVCSEC}
	}
	if !info.requiresFleetAPI {
		c.Flags |= cli.FlagVIN
	}

	// Verify all required parameters are present.
	havePrivateKey := !(c.KeyringKeyName == "" && c.KeyFilename == "")
	haveVIN := c.VIN != ""
	_, err := checkReadiness(commandName, havePrivateKey, haveVIN)
	return err
}

var (
	ErrRequiresVIN        = errors.New("command requires a VIN")
	ErrRequiresPrivateKey = errors.New("command requires a private key")
	ErrUnknownCommand     = errors.New("unrecognized command")
)

func checkReadiness(commandName string, havePrivateKey, haveVIN bool) (*Command, error) {
	info, ok := commands[commandName]
	if !ok {
		return nil, ErrUnknownCommand
	}

	return info, nil
}

func execute(ctx context.Context, acct *account.Account, car *vehicle.Vehicle, args []string) error {
	var err error

	if len(args) == 0 {
		return errors.New("missing COMMAND")
	}
	info, ok := commands[args[0]]
	if !ok {
		return ErrUnknownCommand
	}

	if len(args)-1 < len(info.args) || len(args)-1 > len(info.args)+len(info.optional) {
		writeErr("Invalid number of command line arguments: %d (%d required, %d optional).", len(args), len(info.args), len(info.optional))
		err = ErrCommandLineArgs
	} else {
		keywords := make(map[string]string)
		for i, argInfo := range info.args {
			keywords[argInfo.name] = args[i+1]
		}
		index := len(info.args) + 1
		for _, argInfo := range info.optional {
			if index >= len(args) {
				break
			}
			keywords[argInfo.name] = args[index]
			index++
		}
		err = info.handler(ctx, acct, car, keywords)
	}

	// Print command-specific help
	if errors.Is(err, ErrCommandLineArgs) {
		info.Usage(args[0])
	}
	return err
}

func (c *Command) Usage(name string) {
	fmt.Printf("Usage: %s", name)
	maxLength := 0
	for _, arg := range c.args {
		fmt.Printf(" %s", arg.name)
		if len(arg.name) > maxLength {
			maxLength = len(arg.name)
		}
	}
	if len(c.optional) > 0 {
		fmt.Printf(" [")
	}
	for _, arg := range c.optional {
		fmt.Printf(" %s", arg.name)
		if len(arg.name) > maxLength {
			maxLength = len(arg.name)
		}
	}
	if len(c.optional) > 0 {
		fmt.Printf(" ]")
	}
	fmt.Printf("\n%s\n", c.help)
	maxLength++
	for _, arg := range c.args {
		fmt.Printf("    %s:%s%s\n", arg.name, strings.Repeat(" ", maxLength-len(arg.name)), arg.help)
	}
	for _, arg := range c.optional {
		fmt.Printf("    %s:%s%s\n", arg.name, strings.Repeat(" ", maxLength-len(arg.name)), arg.help)
	}
}

var commands = map[string]*Command{
	"list-keys": &Command{
		help:             "List public keys enrolled on vehicle",
		requiresAuth:     false,
		requiresFleetAPI: false,
		handler: func(ctx context.Context, acct *account.Account, car *vehicle.Vehicle, args map[string]string) error {
			summary, err := car.KeySummary(ctx)
			if err != nil {
				return err
			}
			var keylist []vcsec.WhitelistEntryInfo
			var details *vcsec.WhitelistEntryInfo
			var publicKey []byte
			var keyRole string
			var keyFormFactor string
			slot := uint32(0)
			for mask := summary.GetSlotMask(); mask > 0; mask >>= 1 {
				if mask&1 == 1 {
					details, err = car.KeyInfoBySlot(ctx, slot)
					if err != nil {
						writeErr("Error fetching slot %d: %s", slot, err)
						if errors.Is(err, context.DeadlineExceeded) {
							return err
						}
					}
					if details != nil {
						keylist = append(keylist, *details)
						publicKey = details.GetPublicKey().GetPublicKeyRaw()
						keyRole = fmt.Sprintf("%s", details.GetKeyRole())
						keyFormFactor = fmt.Sprintf("%s", details.GetMetadataForKey().GetKeyFormFactor())
						log.Debug("Key found! public key: %02x, key role: %s, key form factor : %s", publicKey, keyRole, keyFormFactor)
					}
				}
				slot++
			}
			fmt.Printf("\"rssi\":%d,\"keylist\":[", car.RSSI())
			for keyno := range keylist {
				publicKey = keylist[keyno].GetPublicKey().GetPublicKeyRaw()
				keyRole = fmt.Sprintf("%s", keylist[keyno].GetKeyRole())
				keyFormFactor = fmt.Sprintf("%s", keylist[keyno].GetMetadataForKey().GetKeyFormFactor())
				if keyno > 0 {
					fmt.Printf(",")
				}
				fmt.Printf("{\"publicKey\":\"%02x\",\"role\":\"%s\",\"formFactor\":\"%s\"}", publicKey, keyRole, keyFormFactor)
			}
			fmt.Printf("]")
			return nil
		},
	},
	"body-controller-state": &Command{
		help:             "Fetch limited vehicle state information. Works over BLE when infotainment is asleep.",
		domain:           protocol.DomainVCSEC,
		requiresAuth:     false,
		requiresFleetAPI: false,
		optional: []Argument{
			Argument{name: "OUTPUT", help: "'enums' or 'numbers' (default)."},
		},
		handler: func(ctx context.Context, acct *account.Account, car *vehicle.Vehicle, args map[string]string) error {
			var jsondata string
			info, err := car.BodyControllerState(ctx)
			if err != nil {
				return err
			}
			useEnumNumbers := true
			if output, ok := args["OUTPUT"]; ok && strings.ToUpper(output) == "ENUMS" {
				useEnumNumbers = false
			}
			options := protojson.MarshalOptions{
				Multiline:         false,
				Indent:            "",
				UseEnumNumbers:    useEnumNumbers,
				EmitUnpopulated:   false,
				EmitDefaultValues: true,
			}
			jsondata = options.Format(info)
			fmt.Printf("\"rssi\":%d,\"state\":%s", car.RSSI(), jsondata)
			return nil
		},
	},
}

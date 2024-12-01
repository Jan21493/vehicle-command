module github.com/teslamotors/vehicle-command

go 1.23.0

require (
	github.com/99designs/keyring v1.2.2
	github.com/cronokirby/saferith v0.33.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	golang.org/x/term v0.15.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/go-ble/ble v0.0.0-20240122180141-8c5522f54333
	github.com/rigado/ble v0.6.17
)

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/JuulLabs-OSS/cbgo v0.0.1 // indirect
	github.com/aead/cmac v0.0.0-20160719120800-7af84192f0b1 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/jacobsa/go-serial v0.0.0-20180131005756-15cf729a72d4 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mgutz/logxi v0.0.0-20161027140823-aebf8a7d67ab // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/raff/goble v0.0.0-20190909174656-72afc67d6a99 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/t-tomalak/logrus-easy-formatter v0.0.0-20190827215021-c074f06c5816 // indirect
	github.com/wsddn/go-ecdh v0.0.0-20161211032359-48726bab9208 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

replace github.com/JuulLabs-OSS/cbgo => github.com/tinygo-org/cbgo v0.0.4

replace github.com/rigado/ble => /opt/loxberry/rigado/ble

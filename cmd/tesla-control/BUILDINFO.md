
export PATH=$PATH:/opt/loxberry/golang/go/bin
cd ~/vehicle-command/
go get ./...
cd cmd/tesla-control
HWINFO=$(cat /sys/firmware/devicetree/base/model|xargs --null printf "%s")
HWARCH=$(uname -m)
VCVERSION=$(git rev-parse --short HEAD)
TODAY=$(date "+%a, %d %b %Y %T")
echo "Hardware info: "$HWINFO
echo "Hardware architecure: "$HWARCH
echo "Vehicle-Command SDK version: "$VCVERSION
echo "Date: "$TODAY
go build -ldflags "-X 'main.version=$VCVERSION' -X 'main.hwinfo=$HWINFO' -X 'main.hwarch=$HWARCH' -X 'main.today=$TODAY'"  ./...

# use root user for the following two commands
su
mv ./vehicle-command/cmd/tesla-control/tesla-control /usr/local/bin/
setcap 'cap_net_admin=eip' /usr/local/bin/tesla-control

mv ./vehicle-command/cmd/tesla-scan/tesla-scan /usr/local/bin/
setcap 'cap_net_admin=eip' /usr/local/bin/tesla-scan

#!/bin/bash

set -e

ARCH=$(uname -m)
WG_CONF="/etc/wireguard/wg0.conf"
TIMEOUT=30
COUNT=0

if [[ "$ARCH" != "x86_64" ]]; then
    echo "ERROR: This release of Wg-Gen-Plus has only been compiled for 64-bit (x86_64) systems. Detected architecture: $ARCH"
    echo "ARM64 support is coming in a future version"
    exit 1
fi

echo "Installing Wg-Gen-Plus."

apt-get update
apt-get install -y wireguard

mkdir -p /etc/wg-gen-plus
mkdir -p /opt/wg-gen-plus
mkdir -p /opt/wg-api
mkdir -p /var/lib/wg-gen-plus

cp -R ./binaries/x64/* /opt/wg-gen-plus/
cp ./wg-api/wg-api /opt/wg-api/
cp ./files/wg-gen-plus.conf /etc/wg-gen-plus/wg-gen-plus-wg0.conf
cp ./files/wg-gen-plus.service /etc/systemd/system/
cp ./files/wg-api.service /etc/systemd/system/

/usr/bin/systemctl enable wg-gen-plus.service
/usr/bin/systemctl restart wg-gen-plus.service

echo "Waiting for Wg-Gen-Plus to start and complete initial configuration..."
while [ ! -f "$WG_CONF" ]; do
    echo -n "."
    sleep 1
done
echo -e "\nWg-Gen-Plus started!"

/usr/bin/systemctl enable wg-quick@wg0
/usr/bin/systemctl restart wg-quick@wg0

/usr/bin/systemctl enable wg-api.service
/usr/bin/systemctl restart wg-api.service


echo ""
echo "-----------------------------------------------------------"
echo "Wg-Gen-Plus installed and your Wireguard service is enabled"

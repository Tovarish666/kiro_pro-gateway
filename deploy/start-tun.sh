#!/bin/bash
# Setup TUN device and start tun2socks
ip tuntap add dev tun0 mode tun 2>/dev/null || true
ip addr add 198.18.0.1/15 dev tun0 2>/dev/null || true
ip link set dev tun0 up
sleep 1
exec /usr/local/bin/tun2socks -device tun0 -proxy socks5://127.0.0.1:1080 -loglevel warning

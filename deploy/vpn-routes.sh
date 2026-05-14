#!/bin/bash
# Restore VPN routing after tun2socks starts.
# Called by tun2socks.service ExecStartPost.
sleep 3
GW=$(ip route show table main | grep 'default via' | grep ens18 | awk '{print $3; exit}')
if [ -z "$GW" ]; then GW=192.168.88.1; fi
VPN_SERVER=172.86.95.45  # xray VLESS server — must route directly, not via TUN

ip route add "$VPN_SERVER" via "$GW" dev ens18 2>/dev/null || true
ip route replace default dev tun0 2>/dev/null || true

echo "$(date): routes OK, GW=$GW, VPN=$VPN_SERVER" >> /var/log/vpn-routes.log

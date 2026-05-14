#!/usr/bin/env bash
# kiro-proxy full deploy script for Debian/Ubuntu
# Usage: KIRO_API_KEY=ksk_... bash install.sh
set -euo pipefail

KIRO_API_KEY="${KIRO_API_KEY:?Set KIRO_API_KEY=ksk_...}"
INSTALL_DIR=/opt/kiro-proxy
XRAY_VER=$(curl -s https://api.github.com/repos/XTLS/Xray-core/releases/latest | grep tag_name | cut -d'"' -f4)
T2S_VER=$(curl -s https://api.github.com/repos/xjasonlyu/tun2socks/releases/latest | grep tag_name | cut -d'"' -f4)
VPN_SERVER="${VPN_SERVER:-}"  # leave empty to skip VPN

echo "==> Creating directories"
mkdir -p "$INSTALL_DIR/bin" "$INSTALL_DIR/logs"

echo "==> Installing xray $XRAY_VER"
curl -sL "https://github.com/XTLS/Xray-core/releases/download/${XRAY_VER}/Xray-linux-64.zip" -o /tmp/xray.zip
unzip -o /tmp/xray.zip xray geoip.dat geosite.dat -d /usr/local/bin/
chmod +x /usr/local/bin/xray

echo "==> Installing tun2socks $T2S_VER"
curl -sL "https://github.com/xjasonlyu/tun2socks/releases/download/${T2S_VER}/tun2socks-linux-amd64.zip" -o /tmp/t2s.zip
unzip -o /tmp/t2s.zip -d /tmp/t2s/
mv /tmp/t2s/tun2socks-linux-amd64 /usr/local/bin/tun2socks
chmod +x /usr/local/bin/tun2socks

echo "==> Installing kiro-cli"
curl -fsSL https://cli.kiro.dev/install | bash
ln -sf "$HOME/.local/bin/kiro-cli" /usr/local/bin/kiro-cli

echo "==> Deploying kiro-proxy binary"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cp "$SCRIPT_DIR/../bin/kiro-proxy-linux-amd64" "$INSTALL_DIR/bin/kiro-proxy"
chmod +x "$INSTALL_DIR/bin/kiro-proxy"

echo "==> Installing helper scripts"
cp "$SCRIPT_DIR/start-tun.sh" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/vpn-routes.sh" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/"*.sh

echo "==> Installing systemd services"
if [ -f "$SCRIPT_DIR/../configs/xray-client.json" ]; then
    mkdir -p /etc/xray
    echo "   Copy configs/xray-client.json to /etc/xray/config.json and fill in your VLESS credentials"
fi

sed "s|YOUR_KIRO_API_KEY|$KIRO_API_KEY|g" "$SCRIPT_DIR/kiro-proxy.service" \
    > /etc/systemd/system/kiro-proxy.service
cp "$SCRIPT_DIR/xray.service"    /etc/systemd/system/xray.service
cp "$SCRIPT_DIR/tun2socks.service" /etc/systemd/system/tun2socks.service

systemctl daemon-reload
systemctl enable xray tun2socks kiro-proxy

echo ""
echo "==> Done! Next steps:"
echo "  1. Edit /etc/xray/config.json with your VLESS credentials"
echo "  2. systemctl start xray tun2socks kiro-proxy"
echo "  3. curl https://api.ipify.org  # should show VPN IP"
echo "  4. curl http://localhost:8080/healthz"

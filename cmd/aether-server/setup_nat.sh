#!/usr/bin/env bash
# Setup NAT (MASQUERADE) for the AetherL3 tunnel interface.
# Run this script as root after starting aether-server.
#
# Usage: sudo ./setup_nat.sh [tun_interface] [outbound_interface]
# Example: sudo ./setup_nat.sh aether0 eth0

set -euo pipefail

TUN_IF="${1:-aether0}"
OUT_IF="${2:-eth0}"

echo "Setting up NAT: $TUN_IF -> $OUT_IF"

# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1

# Add MASQUERADE rule for IPv4
if ! iptables -t nat -C POSTROUTING -o "$OUT_IF" -j MASQUERADE 2>/dev/null; then
    iptables -t nat -A POSTROUTING -o "$OUT_IF" -j MASQUERADE
    echo "Added IPv4 MASQUERADE rule"
fi

# Allow forwarding from TUN interface
if ! iptables -C FORWARD -i "$TUN_IF" -j ACCEPT 2>/dev/null; then
    iptables -A FORWARD -i "$TUN_IF" -j ACCEPT
fi

if ! iptables -C FORWARD -o "$TUN_IF" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null; then
    iptables -A FORWARD -o "$TUN_IF" -m state --state RELATED,ESTABLISHED -j ACCEPT
fi

# IPv6 NAT (if supported)
if command -v ip6tables &>/dev/null; then
    if ! ip6tables -t nat -C POSTROUTING -o "$OUT_IF" -j MASQUERADE 2>/dev/null; then
        ip6tables -t nat -A POSTROUTING -o "$OUT_IF" -j MASQUERADE
        echo "Added IPv6 MASQUERADE rule"
    fi
fi

echo "NAT setup complete."
echo "Verify with: iptables -t nat -L -n -v"
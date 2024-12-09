FROM registry.cn-hangzhou.aliyuncs.com/117503445-mirror/sync:linux.amd64.ghcr.io.sagernet.sing-box.latest

COPY config.jsonc /etc/sing-box/config.json

# /var/lib/sing-box -C /etc/sing-box/ run
CMD ["-D", "/var/lib/sing-box", "-C", "/etc/sing-box/", "run"]
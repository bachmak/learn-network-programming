package main

import (
	cmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	_ "learn-network-programming/ch10-caddy/restrict_prefix"
	_ "learn-network-programming/ch10-caddy/toml_adapter"
)

func main() {
	cmd.Main()
}

// Command intune is a Linux-style CLI for managing Microsoft Intune resources
// via the Microsoft Graph API. See 'intune --help' for usage.
package main

import (
	"github.com/shumasod/IntuneProvider/internal/cli"
)

// version is injected at build time via:
//
//	go build -ldflags="-X main.version=v0.1.0" ./cmd/intune
var version = "dev"

func main() {
	cli.SetVersion(version)
	cli.Execute()
}

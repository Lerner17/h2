package main

import (
	"fmt"
	"os"

	appconfig "vpn/internal/config"
	"vpn/internal/tui"
)

func main() {
	loadResult, err := appconfig.LoadCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load config: %v\n", err)
		os.Exit(1)
	}
	if loadResult.WasCreated {
		fmt.Fprintf(os.Stderr, "Config created: %s\n", loadResult.Path)
	}

	deps, err := tui.BuildDependencies(loadResult.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := tui.Run(deps); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

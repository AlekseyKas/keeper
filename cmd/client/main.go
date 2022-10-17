//go:build windows || linux || darwin || arm
// +build windows linux darwin arm

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlekseyKas/keeper/pkg/client"
	"github.com/AlekseyKas/keeper/pkg/helpers"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
	_            = flag.Bool("version", false, "Print version information and exit.")
)

func PrintVersion() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Printf("Build Commit: %s\n", buildCommit)
}

func main() {
	flag.Parse()
	if helpers.IsFlagPassed("version") {
		PrintVersion()
		os.Exit(0)
	}
	client := client.New()
	client.Start()
	client.WG().Wait()
}

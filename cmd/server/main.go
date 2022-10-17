package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/server"
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

func init() {
	zerolog.New()
}

func main() {
	flag.Parse()
	if helpers.IsFlagPassed("version") {
		PrintVersion()
		os.Exit(0)
	}
	server := server.New()
	server.Start()
	server.Wg.Wait()
}

package main

import (
	"automateLife/config"
	"automateLife/handlers"
	"automateLife/ui"
	"os"
	"os/user"
)

func main() {
	// Ensure HOME environment variable is set
	if os.Getenv("HOME") == "" {
		if currentUser, err := user.Current(); err == nil {
			os.Setenv("HOME", currentUser.HomeDir)
		}
	}

	fileName := config.DefaultConfigFileName
	args := os.Args

	if len(args) < 2 {
		showHelp()
		return
	}

	switch args[1] {
	case "init":
		handlers.HandleInit(fileName)
	case "start":
		handlers.HandleStart(fileName)
	case "verify":
		handlers.HandleVerify(fileName)
	case "test":
		handlers.HandleTest(fileName)
	case "build":
		handlers.HandleBuild(fileName)
	default:
		showHelp()
	}

}

func showHelp() {
	ui.PrintBanner()
	ui.PrintWelcome()
}

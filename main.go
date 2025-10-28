package main

import (
	"automateLife/config"
	"automateLife/handlers"
	"automateLife/ui"
	"os"
)

func main() {

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
	default:
		showHelp()
	}

}

func showHelp() {
	ui.PrintBanner()
	ui.PrintWelcome()
}

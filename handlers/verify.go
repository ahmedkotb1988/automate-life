package handlers

import (
	"automateLife/config"
	"automateLife/ui"
	"fmt"
)

func HandleVerify(fileName string) {
	cfg, err := config.Load(fileName)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load config: %v", err))
		return
	}

	if err := cfg.Validate(); err != nil {
		ui.Error(fmt.Sprintf("Validation failed: %v", err))
		return
	}

	ui.Success("Directory verified successfully and ready for automation. Run 'automateLife start' to automate!")
}

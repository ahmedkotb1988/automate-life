package handlers

import (
	"automateLife/config"
	"automateLife/ui"
	"fmt"
)

func HandleInit(fileName string) {
	content := config.DefaultConfigTemplate()

	if err := config.Create(fileName, content); err != nil {
		if err.Error() == "config file already exists" {
			fmt.Println(fileName + " already exists in your current directory")
		} else {
			ui.Error(fmt.Sprintf("Failed to create %s: %v", fileName, err))
		}
		return
	}

	ui.Success(fileName + " created successfully")
}

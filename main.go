package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/IbraheemHaseeb7/apee-i/utils"
)

func main() {
	// creating flags to be passed into the program
	file := flag.String("file", "api.json", "file for getting all the api information")
	env := flag.String("env", "development", "environment in which data is to be tested")
	pipeline := flag.String("pipeline", "current", "whether to run current, all custom or selected custom pipeline")
	customPipelineName := flag.String("name", "", "custom pipeline name")
	flag.Parse()

	// generating absolute path for the json file
	filePath, err := filepath.Abs(*file)
	if err != nil { fmt.Println("Could not get absolute path") }

	// reading instructions from the json file
	var fileContents utils.Structure
	err = fileContents.ReadInstructions(filePath)
	if err != nil { fmt.Println(err.Error()); return }

	// creating options for various purposes
	envSelector := map[string]func() {
		"development": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Development; fileContents.ActiveEnvironment = "development" },
		"staging": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Staging; fileContents.ActiveEnvironment = "staging" },
		"production": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Production; fileContents.ActiveEnvironment = "production" },
	}
	pipelineSelector := map[string]interface{} {
		"current": fileContents.CallCurrentPipeline,
		"all": fileContents.CallCustomPipelines,
		"custom": fileContents.CallSingleCustomPipeline,
	}

	// selecting environment and calling selected pipeline by the user
	if action, exists := envSelector[*env]; exists { action()
	} else { fmt.Println("No such environment exists!!!"); return }

	fileContents.Login()
	
	if *pipeline == "custom" { fileContents.CallSingleCustomPipeline(*customPipelineName); return }
	if action, exists := pipelineSelector[*pipeline]; exists { action.(func())()
	} else { fmt.Println("No such pipeline exists!!!"); return }
}


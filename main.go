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
	envSelector := map[string]interface{} {
		"development": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Development},
		"staging": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Staging},
		"production": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Production},
	}
	pipelineSelector := map[string]interface{}{
		"current": fileContents.CallCurrentPipeline,
		"all": fileContents.CallCustomPipelines,
		"custom": fileContents.CallSingleCustomPipeline,
	}

	// selecting environment and calling selected pipeline by the user
	envSelector[*env].(func())()
	fileContents.Login()
	
	if *pipeline == "custom" { fileContents.CallSingleCustomPipeline(*customPipelineName); return }
	pipelineSelector[*pipeline].(func())()
}


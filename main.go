package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/cmd/json"
	"github.com/IbraheemHaseeb7/apee-i/cmd/yaml"
	"github.com/IbraheemHaseeb7/apee-i/utils"
)

func main() {
	// navigating for sub-commands
	if len(os.Args) > 1 {

		subCommand := strings.Split(os.Args[1], "=")[0]
		availableCommands := map[string]func() bool {
			"update": func() bool {
				cmd.Update(); return false
			},
			"": func() bool { return true },
			"-help": func() bool { cmd.Help();return false },
			"--help": func() bool { cmd.Help();return false },
			"-file": func() bool { return true },
			"--file": func() bool { return true },
			"-env": func() bool { return true },
			"--env": func() bool { return true },
			"-pipeline": func() bool { return true },
			"--pipeline": func() bool { return true },
			"-name": func() bool { return true },
			"--name": func() bool { return true },
		}

		if action, exists := availableCommands[subCommand]; exists { if !action() {return};
		} else { fmt.Println("No such subcommand exists!!!\n\n\tTry "+ utils.Green +"apee-i --help"+utils.Reset+" to see all commands\n"); return }
	}

	// creating flags to be passed into the program
	// help := flag.String("help", "", "")
	file := flag.String("file", "api.json", "file for getting all the api information")
	env := flag.String("env", "development", "environment in which data is to be tested")
	pipeline := flag.String("pipeline", "current", "whether to run current, all custom or selected custom pipeline")
	customPipelineName := flag.String("name", "", "custom pipeline name")
	flag.Parse()

	// generating absolute path for the json file
	filePath, err := filepath.Abs(*file)
	if err != nil { fmt.Println("Could not get absolute path") }

	// setting file context to read file and checking for file type
	fileContext := &cmd.FileReaderContext{}
	format := strings.Split(filePath, ".")[1]
	if err != nil { fmt.Println(err.Error()); return }

	// choosing the file reader according to file type
	if format == "yaml" { fileContext.SetStrategy(&yaml.YAMLReader{})
	} else if format == "json" { fileContext.SetStrategy(&json.JSONReader{})
	} else { fmt.Println("Invalid File format!!"); return }

	// calling the instructions reader
	fileContents, err := fileContext.ReadInstructions(filePath)
	if err != nil { fmt.Println("Could not read file"); return }

	// creating options for various purposes
	envSelector := map[string]func() {
		"development": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Development; fileContents.ActiveEnvironment = "development" },
		"staging": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Staging; fileContents.ActiveEnvironment = "staging" },
		"production": func() {fileContents.ActiveUrl = fileContents.BaseUrl.Production; fileContents.ActiveEnvironment = "production" },
	}
	pipelineSelector := map[string]any {
		"current": fileContext.CallCurrentPipeline,
	 	"all": fileContext.CallCustomPipelines,
	 	"custom": fileContext.CallSingleCustomPipeline,
	}
	
	// selecting environment and calling selected pipeline by the user
	if action, exists := envSelector[*env]; exists { action()
	} else { fmt.Println("No such environment exists!!!"); return }

	fileContext.Login(fileContents)
	
	if *pipeline == "custom" { fileContext.CallSingleCustomPipeline(fileContents, *customPipelineName); return }
	if action, exists := pipelineSelector[*pipeline]; exists { action.(func(fileContents *cmd.Structure))(fileContents)
	} else { fmt.Println("No such pipeline exists!!!"); return }
}



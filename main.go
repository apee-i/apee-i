package main

import (
	"flag"
	"fmt"
	"github.com/IbraheemHaseeb7/apee-i/cmd/inline"
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
		availableCommands := map[string]func() bool{
			"update": func() bool {
				cmd.Update()
				return false
			},
			"":           func() bool { return true },
			"-help":      func() bool { cmd.Help(); return false },
			"--help":     func() bool { cmd.Help(); return false },
			"-version":   func() bool { cmd.Version(); return false },
			"--version":  func() bool { cmd.Version(); return false },
			"-file":      func() bool { return true },
			"--file":     func() bool { return true },
			"-env":       func() bool { return true },
			"--env":      func() bool { return true },
			"-pipeline":  func() bool { return true },
			"--pipeline": func() bool { return true },
			"-name":      func() bool { return true },
			"--name":     func() bool { return true },
			"--inline":   func() bool { return true },
		}

		if action, exists := availableCommands[subCommand]; exists {
			if !action() {
				return
			}
		} else {
			fmt.Println("No such subcommand exists!!!\n\n\tTry " + utils.Green + "apee-i --help" + utils.Reset + " to see all commands\n")
			return
		}
	}

	// creating flags to be passed into the program
	// help := flag.String("help", "", "")
	// flags for file based mode
	file := flag.String("file", "api.json", "file for getting all the api information")
	env := flag.String("env", "development", "environment in which data is to be tested")
	pipeline := flag.String("pipeline", "current", "whether to run current, all custom or selected custom pipeline")
	customPipelineName := flag.String("name", "", "custom pipeline name")

	// flags for inline mode
	inlineMode := flag.Bool("inline", false, "Run in inline mode (ignores configuration file)")
	inlineMethod := flag.String("method", "GET", "HTTP method for inline mode (e.g., GET, POST, PUT, DELETE)")
	inlineBody := flag.String("body", "", "Request body as a JSON string for inline mode")
	inlineEndpoint := flag.String("endpoint", "", "Endpoint for inline mode (e.g., v1/users)")
	inlineHeaders := flag.String("headers", "", "Custom header(s) in key:value format (comma-separated, e.g., 'Content-Type:application/json,Accept:application/json')")
	inlineBaseURL := flag.String("url", "", "Base URL for inline mode (e.g., https://api.example.com)")
	inlineProtocol := flag.String("protocol", "HTTP", "Protocol to use for inline mode (currently only HTTP is supported)")
	inlineTimeout := flag.Int("timeout", 30, "Request timeout in seconds for inline mode")
	inlineToken := flag.String("token", "", "Optional authentication token for inline mode")
	inlineExpectedStatusCode := flag.Int("statusCode", 200, "Expected status code for inline mode")

	flag.Parse()

	// inline mode
	if *inlineMode {
		headers := make(map[string]string)

		if *inlineHeaders != "" {
			headerPairs := strings.Split(*inlineHeaders, ",")
			for _, headerPair := range headerPairs {
				kv := strings.Split(headerPair, ":")
				if len(kv) == 2 {
					headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
		}

		if strings.ToUpper(*inlineProtocol) != "HTTP" {
			fmt.Println("Inline mode only supports HTTP protocol at this time.")
			return
		}

		inlineOpts := inline.Options{
			Method:             *inlineMethod,
			Body:               *inlineBody,
			Endpoint:           *inlineEndpoint,
			Headers:            headers,
			BaseURL:            *inlineBaseURL,
			Protocol:           *inlineProtocol,
			Timeout:            *inlineTimeout,
			Token:              *inlineToken,
			ExpectedStatusCode: *inlineExpectedStatusCode,
		}

		inline.RunInline(inlineOpts)
		return
	}

	// file based mode
	// generating absolute path for the json file
	filePath, err := filepath.Abs(*file)
	if err != nil {
		fmt.Println("Could not get absolute path")
	}

	// setting file context to read file and checking for file type
	fileContext := &cmd.FileReaderContext{}
	format := strings.Split(filePath, ".")[1]
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// choosing the file reader according to file type
	if format == "yaml" {
		fileContext.SetStrategy(&yaml.Strategy{})
	} else if format == "json" {
		fileContext.SetStrategy(&json.Strategy{})
	} else {
		fmt.Println("Invalid File format!!")
		return
	}

	// calling the instructions reader
	fileContents, err := fileContext.ReadInstructions(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// creating options for various purposes
	envSelector := map[string]func(){
		"development": func() {
			fileContents.ActiveURL = fileContents.BaseURL.Development
			fileContents.ActiveEnvironment = "development"
		},
		"staging": func() {
			fileContents.ActiveURL = fileContents.BaseURL.Staging
			fileContents.ActiveEnvironment = "staging"
		},
		"production": func() {
			fileContents.ActiveURL = fileContents.BaseURL.Production
			fileContents.ActiveEnvironment = "production"
		},
	}
	pipelineSelector := map[string]any{
		"current": fileContext.CallCurrentPipeline,
		"all":     fileContext.CallCustomPipelines,
		"custom":  fileContext.CallSingleCustomPipeline,
	}

	// selecting environment and calling selected pipeline by the user
	if action, exists := envSelector[*env]; exists {
		action()
	} else {
		fmt.Println("No such environment exists!!!")
		return
	}

	fileContext.Login(fileContents)

	if *pipeline == "custom" {
		fileContext.CallSingleCustomPipeline(fileContents, *customPipelineName)
		return
	}
	if action, exists := pipelineSelector[*pipeline]; exists {
		action.(func(fileContents *cmd.Structure))(fileContents)
	} else {
		fmt.Println("No such pipeline exists!!!")
		return
	}
}

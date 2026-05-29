package yaml

import (
	"fmt"
	"os"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/cmd/protocols"
	myHttp "github.com/IbraheemHaseeb7/apee-i/cmd/protocols/http"
	"github.com/IbraheemHaseeb7/apee-i/cmd/protocols/ws"
	"github.com/IbraheemHaseeb7/apee-i/utils"
)

// Login is use to login user based on login credentials
// provided with YAML file. Checks for the environment and
// then uses credentials accordingly
func (r *Strategy) Login(fileContents *cmd.Structure) {

	fmt.Print("\t ==> Performing Authentication <==\n\n")
	fmt.Println(utils.White + "* Looking for token 👀" + utils.Reset)
	// checking if token exists in the file
	// if file doesnt exist, generate new token
	data, err := os.ReadFile("token.txt")
	if err != nil {
		fmt.Println(utils.White + "* Token file not found ❌" + utils.Reset)
		GetAndStoreToken(fileContents)
		return
	}

	// if file exists but is empty, generate new token
	if string(data) == "" {
		fmt.Println(utils.White + "* Token not present in the document ❌" + utils.Reset)
		GetAndStoreToken(fileContents)
		return
	}

	// store token in app state
	fileContents.LoginDetails.Token = string(data)

	fmt.Println(utils.White + "* Token found ✅" + utils.Reset)
	fmt.Print(utils.White + "* Testing for valid token 🔃" + utils.Reset + "\n\n")

	pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
		Endpoint: fileContents.LoginDetails.TestingRoute,
	})

	protocolContext := &protocols.ProtocolContext{}
	protocolContext.SetStrategy(&myHttp.Strategy{})

	tokenCheckResponse, err := protocolContext.Hit(fileContents, *pipeline)
	if err != nil {
		fmt.Print(utils.White + "\n* Could not hit API, try again ❌" + utils.Reset + "\n\n")
		return
	}

	// if request fails with unauthorized, generate new token
	if tokenCheckResponse.StatusCode != 200 && tokenCheckResponse.StatusCode != 201 {
		fmt.Println(utils.White + "\n* Invalid token found ❌" + utils.Reset)
		GetAndStoreToken(fileContents)
		return
	}

	fmt.Println(utils.White + "\nValid token found ✅\n" + utils.Reset)
}

// GetAndStoreToken is a helper function that simply gets the token
// from the response and store it into token.txt file
func GetAndStoreToken(fileContents *cmd.Structure) {

	credentials := fileContents.Credentials.Development
	if fileContents.ActiveEnvironment == "development" {
		credentials = fileContents.Credentials.Development
	}
	if fileContents.ActiveEnvironment == "staging" {
		credentials = fileContents.Credentials.Staging
	}
	if fileContents.ActiveEnvironment == "production" {
		credentials = fileContents.Credentials.Production
	}

	fmt.Print(utils.White + "* Generating and storing new token ⚙️" + utils.Reset + "\n\n")

	protocolContext := &protocols.ProtocolContext{}
	protocolContext.SetStrategy(&myHttp.Strategy{})

	// hitting login api with credentials
	loginDetails := cmd.NewLoginDetails(&fileContents.LoginDetails)
	tokenGetResponse, err := protocolContext.Hit(fileContents, *cmd.NewPipelineBody(&cmd.PipelineBody{
		Endpoint:           loginDetails.Route,
		Method:             "POST",
		Body:               credentials.Body,
		Headers:            credentials.Headers,
		ExpectedStatusCode: credentials.ExpectedStatusCode,
	}))
	if err != nil {
		fmt.Println(utils.White + "* Could not hit API, try again ❌" + utils.Reset)
		return
	}

	// fetching token form the json response from the given structure in json file
	token, _ := tokenGetResponse.Body.Path(fileContents.LoginDetails.TokenLocation).Data().(string)

	// storing token in the file and in app state
	os.WriteFile("token.txt", []byte(token), 0633)
	fileContents.LoginDetails.Token = token
	fmt.Println()
}

// CallCurrentPipeline calls the current pipeline APIs endpoints in a sequence
func (r *Strategy) CallCurrentPipeline(fileContents *cmd.Structure) {

	fmt.Print("\t ==> Calling All APIs in \"Current\" Pipeline <==\n\n")
	for i := range fileContents.CurrentPipeline.Pipeline {
		mergedHeaders := utils.Merge2Maps(fileContents.CurrentPipeline.Globals.Headers, fileContents.CurrentPipeline.Pipeline[i].Headers)

		pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
			Endpoint:           fileContents.CurrentPipeline.Pipeline[i].Endpoint,
			Method:             fileContents.CurrentPipeline.Pipeline[i].Method,
			Body:               fileContents.CurrentPipeline.Pipeline[i].Body,
			ExpectedStatusCode: fileContents.CurrentPipeline.Pipeline[i].ExpectedStatusCode,
			Headers:            mergedHeaders,
			BaseURL:            fileContents.CurrentPipeline.Pipeline[i].BaseURL,
			Timeout:            fileContents.CurrentPipeline.Pipeline[i].Timeout,
			Protocol:           fileContents.CurrentPipeline.Pipeline[i].Protocol,
		})

		protocolContext := &protocols.ProtocolContext{}
		if pipeline.Protocol == "HTTP" {
			protocolContext.SetStrategy(&myHttp.Strategy{})
		} else if pipeline.Protocol == "WS" {
			protocolContext.SetStrategy(&ws.Strategy{})
		} else {
			fmt.Print(utils.White + "* Not a valid protocol ❌" + utils.Reset + "\n\n")
			return
		}

		res, err := protocolContext.Hit(fileContents, *pipeline)
		if err != nil {
			fmt.Print(utils.White + "* Could not hit API, try again with different values ❌" + utils.Reset + "\n\n")
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

// CallCustomPipelines calls all the custom pipelines APIs endpoints
func (r *Strategy) CallCustomPipelines(fileContents *cmd.Structure) {

	fmt.Print("\t ==> Calling All APIs in \"Custom\" Pipelines <==\n\n")
	for _, value := range fileContents.CustomPipelines.Pipeline {
		structure := value.([]any)
		for i := range structure {

			req := structure[i].(map[string]any)

			if req["baseUrl"] == nil {
				req["baseUrl"] = ""
			}
			if req["timeout"] == nil {
				req["timeout"] = 0
			}
			if req["protocol"] == nil {
				req["protocol"] = ""
			}
			if req["expectedStatusCode"] == nil {
				req["expectedStatusCode"] = 0
			}
			if req["headers"] == nil {
				req["headers"] = make(map[string]any)
			}
			if fileContents.CustomPipelines.Globals.Headers == nil {
				fileContents.CustomPipelines.Globals.Headers = make(map[string]any)
			}

			mergedHeaders := utils.Merge2Maps(req["headers"].(map[string]any), fileContents.CustomPipelines.Globals.Headers)

			pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
				Endpoint:           req["endpoint"].(string),
				Method:             req["method"].(string),
				Body:               req["body"],
				ExpectedStatusCode: req["expectedStatusCode"].(int),
				Headers:            mergedHeaders,
				BaseURL:            req["baseUrl"].(string),
				Timeout:            req["timeout"].(int),
				Protocol:           req["protocol"].(string),
			})

			protocolContext := &protocols.ProtocolContext{}
			if pipeline.Protocol == "HTTP" {
				protocolContext.SetStrategy(&myHttp.Strategy{})
			} else if pipeline.Protocol == "WS" {
				protocolContext.SetStrategy(&ws.Strategy{})
			} else {
				fmt.Print(utils.Red + "* Not a valid protocol ❌" + utils.Reset + "\n\n")
				return
			}

			res, err := protocolContext.Hit(fileContents, *pipeline)
			if err != nil {
				fmt.Print(utils.White + "* Could not hit API, try again with different values ❌" + utils.Reset + "\n\n")
				return
			}

			fmt.Println(res.Body.StringIndent("", "  "))
		}
	}
}

// CallSingleCustomPipeline calls a single custom pipeline in a sequence
func (r *Strategy) CallSingleCustomPipeline(fileContents *cmd.Structure, pipelineKey string) {

	if fileContents.CustomPipelines.Pipeline[pipelineKey] == nil {
		fmt.Print(utils.White + "* Could not find this pipeline, try another name? ❌" + utils.Reset + "\n\n")
		return
	}
	data := fileContents.CustomPipelines.Pipeline[pipelineKey].([]any)
	fmt.Print("\t ==> Calling \"" + pipelineKey + "\" APIs in \"Custom\" Pipelines <==\n\n")

	for i := range data {
		req := data[i].(map[string]any)

		if req["baseUrl"] == nil {
			req["baseUrl"] = ""
		}
		if req["timeout"] == nil {
			req["timeout"] = 0
		}
		if req["protocol"] == nil {
			req["protocol"] = ""
		}
		if req["expectedStatusCode"] == nil {
			req["expectedStatusCode"] = 0
		}
		if req["headers"] == nil {
			req["headers"] = make(map[string]any)
		}
		if fileContents.CustomPipelines.Globals.Headers == nil {
			fileContents.CustomPipelines.Globals.Headers = make(map[string]any)
		}
		mergedHeaders := utils.Merge2Maps(req["headers"].(map[string]any), fileContents.CustomPipelines.Globals.Headers)

		pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
			Endpoint:           req["endpoint"].(string),
			Method:             req["method"].(string),
			Body:               req["body"],
			ExpectedStatusCode: req["expectedStatusCode"].(int),
			Headers:            mergedHeaders,
			BaseURL:            req["baseUrl"].(string),
			Timeout:            req["timeout"].(int),
			Protocol:           req["protocol"].(string),
		})

		protocolContext := &protocols.ProtocolContext{}
		if pipeline.Protocol == "HTTP" {
			protocolContext.SetStrategy(&myHttp.Strategy{})
		} else if pipeline.Protocol == "WS" {
			protocolContext.SetStrategy(&ws.Strategy{})
		} else {
			fmt.Print(utils.White + "* Not a valid protocol ❌" + utils.Reset + "\n\n")
			return
		}

		res, err := protocolContext.Hit(fileContents, *pipeline)
		if err != nil {
			fmt.Print(utils.White + "* Could not hit API, try again with different values ❌" + utils.Reset + "\n\n")
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

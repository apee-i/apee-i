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

	fmt.Println(utils.Green + "- Looking for token..." + utils.Reset)
	// checking if token exists in the file
	// if file doesnt exist, generate new token
	data, err := os.ReadFile("token.txt")
	if err != nil {
		fmt.Println(utils.Red + "- Token file not found..." + utils.Reset)
		GetAndStoreToken(fileContents)
		return
	}

	// if file exists but is empty, generate new token
	if string(data) == "" {
		fmt.Println(utils.Red + "- Token not present in the document..." + utils.Reset)
		GetAndStoreToken(fileContents)
		return
	}

	// store token in app state
	fileContents.LoginDetails.Token = string(data)

	// getting data from /me api
	fmt.Println(utils.Blue + "- Token found..." + utils.Reset)
	fmt.Println(utils.Blue + "- Testing for valid token..." + utils.Reset)

	pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
		Endpoint: "/me",
	})

	protocolContext := &protocols.ProtocolContext{}
	protocolContext.SetStrategy(&myHttp.Strategy{})

	tokenCheckResponse, err := protocolContext.Hit(fileContents, *pipeline)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
		return
	}

	// if request fails with unauthorized, generate new token
	if tokenCheckResponse.StatusCode == 401 {
		fmt.Println(utils.Red + "- Invalid token found..." + utils.Reset)
		GetAndStoreToken(fileContents)
		return
	}

	fmt.Println(utils.Green + "\nValid token found!!\n" + utils.Reset)
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

	fmt.Println(utils.Green + "- Generating and storing new token..." + utils.Reset)

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
		fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
		return
	}

	// fetching token form the json response from the given structure in json file
	token, _ := tokenGetResponse.Body.Path(fileContents.LoginDetails.TokenLocation).Data().(string)

	// storing token in the file and in app state
	os.WriteFile("token.txt", []byte(token), 0633)
	fileContents.LoginDetails.Token = token
}

// CallCurrentPipeline calls the current pipeline APIs endpoints in a sequence
func (r *Strategy) CallCurrentPipeline(fileContents *cmd.Structure) {

	fmt.Println(utils.Blue + "\nCalling All API in current pipeline\n" + utils.Reset)
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
			fmt.Println(utils.Red + "Not a valid protocol" + utils.Reset)
			return
		}

		res, err := protocolContext.Hit(fileContents, *pipeline)
		if err != nil {
			fmt.Println(utils.Red + err.Error() + utils.Reset)
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

// CallCustomPipelines calls all the custom pipelines APIs endpoints
func (r *Strategy) CallCustomPipelines(fileContents *cmd.Structure) {

	for _, value := range fileContents.CustomPipelines.Pipeline {
		structure := value.([]any)
		for i := range structure {

			req := structure[i].(map[string]any)
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
				fmt.Println(utils.Red + "Not a valid protocol" + utils.Reset)
				return
			}

			res, err := protocolContext.Hit(fileContents, *pipeline)
			if err != nil {
				fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
				return
			}

			fmt.Println(res.Body.StringIndent("", "  "))
		}
	}
}

// CallSingleCustomPipeline calls a single custom pipeline in a sequence
func (r *Strategy) CallSingleCustomPipeline(fileContents *cmd.Structure, pipelineKey string) {
	data := fileContents.CustomPipelines.Pipeline[pipelineKey].([]any)

	for i := range data {
		req := data[i].(map[string]any)
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
			fmt.Println(utils.Red + "Not a valid protocol" + utils.Reset)
			return
		}

		res, err := protocolContext.Hit(fileContents, *pipeline)
		if err != nil {
			fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

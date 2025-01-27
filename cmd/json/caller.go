package json

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/cmd/protocols"
	myHttp "github.com/IbraheemHaseeb7/apee-i/cmd/protocols/http"
	"github.com/IbraheemHaseeb7/apee-i/cmd/protocols/ws"
	"github.com/IbraheemHaseeb7/apee-i/utils"
	"github.com/Jeffail/gabs/v2"
)

// Login function logs the user in based on the credentials
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

	protocolContext := &protocols.ProtocolContext{}
	protocolContext.SetStrategy(&myHttp.Strategy{})

	tokenCheckResponse, err := protocolContext.Hit(fileContents, *cmd.NewPipelineBody(&cmd.PipelineBody{
		Endpoint: "/me",
	}))
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

	// hitting login api with credentials
	loginDetails := cmd.NewLoginDetails(&fileContents.LoginDetails)
	pipelineBody := cmd.NewPipelineBody(&cmd.PipelineBody{
		Endpoint:           loginDetails.Route,
		Method:             "POST",
		Body:               credentials.Body,
		Headers:            credentials.Headers,
		ExpectedStatusCode: credentials.ExpectedStatusCode,
	})

	protocolContext := &protocols.ProtocolContext{}
	protocolContext.SetStrategy(&myHttp.Strategy{})

	tokenGetResponse, err := protocolContext.Hit(fileContents, *pipelineBody)
	if err != nil {
		fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
		return
	}

	// fetching token form the json response from the given structure in json file
	token, _ := tokenGetResponse.Body.Path(loginDetails.TokenLocation).Data().(string)

	// storing token in the file and in app state
	os.WriteFile("token.txt", []byte(token), 0633)
	fileContents.LoginDetails.Token = token
}

// CallCurrentPipeline calls the current pipeline APIs endpoints in a sequence
func (r *Strategy) CallCurrentPipeline(fileContents *cmd.Structure) {

	mergedHeaders := utils.Merge2Maps(fileContents.CurrentPipeline.Globals.Headers, fileContents.CurrentPipeline.Pipeline[0].Headers)

	fmt.Println(utils.Blue + "\nCalling All API in current pipeline\n" + utils.Reset)
	for i := range fileContents.CurrentPipeline.Pipeline {

		pipelineBody := *cmd.NewPipelineBody(&cmd.PipelineBody{
			Endpoint:           fileContents.CurrentPipeline.Pipeline[i].Endpoint,
			Method:             fileContents.CurrentPipeline.Pipeline[i].Method,
			Body:               fileContents.CurrentPipeline.Pipeline[i].Body,
			ExpectedStatusCode: fileContents.CurrentPipeline.Pipeline[i].ExpectedStatusCode,
			BaseURL:            fileContents.CurrentPipeline.Pipeline[i].BaseURL,
			Headers:            mergedHeaders,
			Timeout:            fileContents.CurrentPipeline.Pipeline[i].Timeout,
			Protocol:           fileContents.CurrentPipeline.Pipeline[i].Protocol,
		})

		protocolContext := &protocols.ProtocolContext{}
		if pipelineBody.Protocol == "HTTP" {
			protocolContext.SetStrategy(&myHttp.Strategy{})
		} else if pipelineBody.Protocol == "WS" {
			protocolContext.SetStrategy(&ws.Strategy{})
		} else {
			fmt.Println(utils.Red + "Not a valid protocol" + utils.Reset)
			return
		}

		res, err := protocolContext.Hit(fileContents, pipelineBody)
		if err != nil {
			fmt.Println(utils.Red + err.Error() + utils.Reset)
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

// CallCustomPipelines calls all the custom pipelines APIs endpoints
func (r *Strategy) CallCustomPipelines(fileContents *cmd.Structure) {
	bytesData, err := json.Marshal(fileContents.CustomPipelines)
	jsonObj, err := gabs.ParseJSON(bytesData)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	children := jsonObj.ChildrenMap()

	for _, value := range children {
		for i := range value.Children() {
			data := value.Children()[i].ChildrenMap()

			mergedHeaders := utils.Merge2Maps(fileContents.CustomPipelines.Globals.Headers, data["headers"].Data().(map[string]any))
			pipelineBody := *cmd.NewPipelineBody(&cmd.PipelineBody{
				Endpoint:           data["endpoint"].Data().(string),
				Method:             data["method"].Data().(string),
				Body:               data["body"].Data(),
				ExpectedStatusCode: int(data["expectedStatusCode"].Data().(float64)),
				BaseURL:            data["baseUrl"].Data().(string),
				Headers:            mergedHeaders,
				Timeout:            data["timeout"].Data().(int),
				Protocol:           data["protocol"].Data().(string),
			})

			protocolContext := &protocols.ProtocolContext{}
			if pipelineBody.Protocol == "HTTP" {
				protocolContext.SetStrategy(&myHttp.Strategy{})
			} else if pipelineBody.Protocol == "WS" {
				protocolContext.SetStrategy(&ws.Strategy{})
			} else {
				fmt.Println(utils.Red + "Not a valid protocol" + utils.Reset)
				return
			}

			res, err := protocolContext.Hit(fileContents, pipelineBody)
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

	bytesData, err := json.Marshal(fileContents.CustomPipelines)
	jsonObj, err := gabs.ParseJSON(bytesData)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	children := jsonObj.ChildrenMap()

	for i := range children[pipelineKey].Children() {
		data := children[pipelineKey].Children()[i].ChildrenMap()

		mergedHeaders := utils.Merge2Maps(fileContents.CustomPipelines.Globals.Headers, data["headers"].Data().(map[string]any))
		pipelineBody := *cmd.NewPipelineBody(&cmd.PipelineBody{
			Endpoint:           data["endpoint"].Data().(string),
			Method:             data["method"].Data().(string),
			Body:               data["body"].Data(),
			ExpectedStatusCode: int(data["expectedStatusCode"].Data().(float64)),
			BaseURL:            data["baseUrl"].Data().(string),
			Headers:            mergedHeaders,
			Timeout:            data["timeout"].Data().(int),
			Protocol:           data["protocol"].Data().(string),
		})

		protocolContext := &protocols.ProtocolContext{}
		if pipelineBody.Protocol == "HTTP" {
			protocolContext.SetStrategy(&myHttp.Strategy{})
		} else if pipelineBody.Protocol == "WS" {
			protocolContext.SetStrategy(&ws.Strategy{})
		} else {
			fmt.Println(utils.Red + "Not a valid protocol" + utils.Reset)
			return
		}

		res, err := protocolContext.Hit(fileContents, pipelineBody)
		if err != nil {
			fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

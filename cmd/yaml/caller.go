package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/utils"
	"github.com/Jeffail/gabs/v2"
)

// Hit acts as an HTTP client and hits a rest based request
func Hit(fileContents *cmd.Structure, structure cmd.PipelineBody) (cmd.APIResponse, error) {

	startTime := time.Now()

	// forming complete url with endpoint
	url := ""
	if structure.BaseURL != "" {
		url = structure.BaseURL + structure.Endpoint
	} else {
		url = fileContents.ActiveURL + structure.Endpoint
	}

	// forming request body for login in json format
	jsonCredentials, err := json.Marshal(structure.Body)
	if err != nil {
		return cmd.APIResponse{}, err
	}

	// forming HTTP request
	req, err := http.NewRequest(structure.Method, url, bytes.NewBuffer(jsonCredentials))
	if err != nil {
		return cmd.APIResponse{}, err
	}

	// adding appropriate headers
	req.Header.Set("Authorization", "bearer"+fileContents.LoginDetails.Token)

	// adding custom headers from the user
	if structure.Headers != nil {
		for key, value := range structure.Headers {
			req.Header.Set(key, value.(string))
		}
	}

	// hitting the server with the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return cmd.APIResponse{}, err
	}

	// closing body when function is popped from stack
	defer res.Body.Close()

	// reading the body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return cmd.APIResponse{}, err
	}

	// parsing body into nice json
	data, err := gabs.ParseJSON(body)
	if err != nil {
		return cmd.APIResponse{}, err
	}

	elapsedTime := time.Since(startTime)

	// logging result - function stored in `helper.go`
	utils.ResponseLogger(structure, res, url, elapsedTime)

	// returning response
	return cmd.APIResponse{
		StatusCode: res.StatusCode,
		Body:       data,
	}, nil
}

// Login is use to login user based on login credentials
// provided with YAML file. Checks for the environment and
// then uses credentials accordingly
func (r *Reader) Login(fileContents *cmd.Structure) {

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
	tokenCheckResponse, err := Hit(fileContents, *pipeline)
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
	tokenGetResponse, err := Hit(fileContents, *cmd.NewPipelineBody(&cmd.PipelineBody{
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
func (r *Reader) CallCurrentPipeline(fileContents *cmd.Structure) {

	fmt.Println(utils.Blue + "\nCalling All API in current pipeline\n" + utils.Reset)
	for i := range fileContents.CurrentPipeline.Pipeline {
		pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
			Endpoint:           fileContents.CurrentPipeline.Pipeline[i].Endpoint,
			Method:             fileContents.CurrentPipeline.Pipeline[i].Method,
			Body:               fileContents.CurrentPipeline.Pipeline[i].Body,
			ExpectedStatusCode: fileContents.CurrentPipeline.Pipeline[i].ExpectedStatusCode,
			Headers:            fileContents.CurrentPipeline.Pipeline[i].Headers,
			BaseURL:            fileContents.CurrentPipeline.Pipeline[i].BaseURL,
		})
		res, err := Hit(fileContents, *pipeline)
		if err != nil {
			fmt.Println(utils.Red + err.Error() + utils.Reset)
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

// CallCustomPipelines calls all the custom pipelines APIs endpoints
func (r *Reader) CallCustomPipelines(fileContents *cmd.Structure) {

	for _, value := range fileContents.CustomPipelines.Pipeline {
		structure := value.([]any)
		for i := range structure {

			req := structure[i].(map[string]any)

			pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
				Endpoint:           req["endpoint"].(string),
				Method:             req["method"].(string),
				Body:               req["body"],
				ExpectedStatusCode: req["expectedStatusCode"].(int),
				Headers:            req["headers"].(map[string]any),
				BaseURL:            req["baseUrl"].(string),
			})
			res, err := Hit(fileContents, *pipeline)
			if err != nil {
				fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
				return
			}

			fmt.Println(res.Body.StringIndent("", "  "))
		}
	}
}

// CallSingleCustomPipeline calls a single custom pipeline in a sequence
func (r *Reader) CallSingleCustomPipeline(fileContents *cmd.Structure, pipelineKey string) {
	data := fileContents.CustomPipelines.Pipeline[pipelineKey].([]any)

	for i := range data {
		req := data[i].(map[string]any)

		pipeline := cmd.NewPipelineBody(&cmd.PipelineBody{
			Endpoint:           req["endpoint"].(string),
			Method:             req["method"].(string),
			Body:               req["body"],
			ExpectedStatusCode: req["expectedStatusCode"].(int),
			Headers:            req["headers"].(map[string]any),
			BaseURL:            req["baseUrl"].(string),
		})
		res, err := Hit(fileContents, *pipeline)
		if err != nil {
			fmt.Println(utils.Red + "Could not hit API, try again..." + utils.Reset)
			return
		}

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

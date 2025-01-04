package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type ApiStructure struct {
	Method string
	Endpoint string 
	Body any
	ExpectedStatusCode int
	ExpectedBody any
	Headers any
}

type ApiResponse struct {
	StatusCode int
	Body *gabs.Container
}

func (fileContents Structure) Hit(structure ApiStructure) (ApiResponse, error) {

	startTime := time.Now()
	if structure.Method == "" { structure.Method = "GET" }

	// forming complete url with endpoint
	url := fileContents.ActiveUrl + structure.Endpoint

	// forming request body for login in json format
	jsonCredentials, err := json.Marshal(structure.Body)
	if err != nil { return ApiResponse{}, err }

	// forming HTTP request
	req, err := http.NewRequest(structure.Method, url, bytes.NewBuffer(jsonCredentials))
	if err != nil { return ApiResponse{}, err }

	// adding appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer" + fileContents.LoginDetails.Token)

	// adding custom headers from the user
	if structure.Headers != nil {
		for key, value := range structure.Headers.(map[string]interface{}) {
			req.Header.Set(key, value.(string))
		}
	}

	// hitting the server with the request
	res, err := http.DefaultClient.Do(req)
	if err != nil { return ApiResponse{}, err }

	// closing body when function is popped from stack
	defer res.Body.Close()

	// reading the body
	body, err := io.ReadAll(res.Body)
	if err != nil { return ApiResponse{}, err }

	// parsing body into nice json
	data, err := gabs.ParseJSON(body)
	if err != nil { return ApiResponse{}, err }

	elapsedTime := time.Since(startTime)

	// logging result - function stored in `helper.go`
	ResponseLogger(structure, res, url, elapsedTime)
	
	// returning response
	return ApiResponse{
		StatusCode: res.StatusCode,
		Body: data,
	}, nil
}

func (fileContents* Structure) Login() {

	fmt.Println(Green + "- Looking for token..." + Reset)
	// checking if token exists in the file 
	// if file doesnt exist, generate new token
	data, err := os.ReadFile("token.txt")
	if err != nil { 
		fmt.Println(Red + "- Token file not found..." + Reset)
		GetAndStoreToken(fileContents); return
	}

	// if file exists but is empty, generate new token
	if string(data) == "" { 
		fmt.Println(Red + "- Token not present in the document..." + Reset)
		GetAndStoreToken(fileContents); return 
	}

	// store token in app state
	fileContents.LoginDetails.Token = string(data)

	// getting data from /me api
	fmt.Println(Blue + "- Token found..." + Reset)
	fmt.Println(Blue + "- Testing for valid token..." + Reset)
	tokenCheckResponse, err:= fileContents.Hit(ApiStructure{
		Endpoint: "/me",
	})
	if err != nil {fmt.Println(err.Error()); fmt.Println(Red + "Could not hit API, try again..." + Reset); return }

	// if request fails with unauthorized, generate new token
	if tokenCheckResponse.StatusCode == 401 {
		fmt.Println(Red + "- Invalid token found..." + Reset)
		GetAndStoreToken(fileContents)
		return 
	}

	fmt.Println(Green + "\nValid token found!!\n" + Reset)
}

func GetAndStoreToken(fileContents* Structure) {
	
	credentials := fileContents.Credentials.Development
	if fileContents.ActiveEnvironment == "development" { credentials = fileContents.Credentials.Development }
	if fileContents.ActiveEnvironment == "staging" { credentials = fileContents.Credentials.Staging }
	if fileContents.ActiveEnvironment == "production" { credentials = fileContents.Credentials.Production }

	fmt.Println(Green + "- Generating and storing new token..." + Reset)
	// hitting login api with credentials
	tokenGetResponse, err := fileContents.Hit(ApiStructure{
		Endpoint: "/login",
		Method: "POST",
		Body: credentials,
	})
	if err != nil { fmt.Println(Red + "Could not hit API, try again..." + Reset); return }

	// fetching token form the json response from the given structure in json file
	token, _ := tokenGetResponse.Body.Path(fileContents.LoginDetails.TokenLocation).Data().(string)

	// storing token in the file and in app state
	os.WriteFile("token.txt", []byte(token), 0633)
	fileContents.LoginDetails.Token = token
}

func (fileContents* Structure) CallCurrentPipeline() {

	fmt.Println(Blue + "\nCalling All API in current pipeline\n" + Reset)
	for i := range fileContents.PipelineBody {
		res, err := fileContents.Hit(ApiStructure{
			Endpoint: fileContents.PipelineBody[i].Endpoint,
			Method: fileContents.PipelineBody[i].Method,
			Body: fileContents.PipelineBody[i].Body,
			ExpectedStatusCode: fileContents.PipelineBody[i].ExpectedStatusCode,
			Headers: fileContents.PipelineBody[i].Headers,
		})
		if err != nil { fmt.Println(Red + "Could not hit API, try again..." + Reset); return }

		fmt.Println(res.Body.StringIndent("", "  "))
	}
}

func (fileContents* Structure) CallCustomPipelines() {
	bytesData, err := json.Marshal(fileContents.CustomPipelines)
	jsonObj, err := gabs.ParseJSON(bytesData)
	if err != nil { fmt.Println(err.Error()); return }

	children:= jsonObj.ChildrenMap()
	
	for _ , value := range children {
		for i := range value.Children() {
			data := value.Children()[i].ChildrenMap()

			if data["method"].Data() == nil { data["method"] = gabs.Wrap("GET") }
			if data["expectedStatusCode"].Data() == nil { 
				var expectedStatusCode float64; data["expectedStatusCode"] = gabs.Wrap(expectedStatusCode)
			}

			res, err := fileContents.Hit(ApiStructure{
				Endpoint: data["endpoint"].Data().(string),
				Method: data["method"].Data().(string),
				Body: data["body"].Data(),
				ExpectedStatusCode: int(data["expectedStatusCode"].Data().(float64)),
				Headers: data["headers"].Data(),
			})
			if err != nil { fmt.Println(Red + "Could not hit API, try again..." + Reset); return }

			fmt.Println(res.Body.StringIndent("", "  "))
		}
	}

}

func (fileContents* Structure) CallSingleCustomPipeline(pipelineKey string) {

	bytesData, err := json.Marshal(fileContents.CustomPipelines)
	jsonObj, err := gabs.ParseJSON(bytesData)
	if err != nil { fmt.Println(err.Error()); return }

	children:= jsonObj.ChildrenMap()
	
	for key, value := range children {
		if key == pipelineKey {
			for i := range value.Children() {
				data := value.Children()[i].ChildrenMap()

				if data["method"].Data() == nil { data["method"] = gabs.Wrap("GET") }
				if data["expectedStatusCode"].Data() == nil { 
					var expectedStatusCode float64; data["expectedStatusCode"] = gabs.Wrap(expectedStatusCode)
				}

				res, err := fileContents.Hit(ApiStructure{
					Endpoint: data["endpoint"].Data().(string),
					Method: data["method"].Data().(string),
					Body: data["body"].Data(),
					ExpectedStatusCode: int(data["expectedStatusCode"].Data().(float64)),
					Headers: data["headers"].Data(),
				})
				if err != nil { fmt.Println(Red + "Could not hit API, try again..." + Reset); return }

				fmt.Println(res.Body.StringIndent("", "  "))
			}
		}
	}
}

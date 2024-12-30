package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type ApiStructure struct {
	Method string `default:"GET"`
	Endpoint string 
	Body any
}

type ApiResponse struct {
	StatusCode int
	Body *gabs.Container
}

func (fileContents Structure) Hit(structure ApiStructure) ApiResponse {

	startTime := time.Now()
	if structure.Method == "" { structure.Method = "GET" }

	// forming complete url with endpoint
	url := fileContents.ActiveUrl + structure.Endpoint

	// forming request body for login in json format
	jsonCredentials, err := json.Marshal(structure.Body)
	if err != nil { fmt.Println("Could not convert into json") }

	// forming HTTP request
	req, err := http.NewRequest(structure.Method, url, bytes.NewBuffer(jsonCredentials))
	if err != nil { fmt.Println(err.Error()) }

	// adding appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer" + fileContents.LoginDetails.Token)

	// hitting the server with the request
	res, err := http.DefaultClient.Do(req)
	if err != nil { fmt.Println(err.Error()) }

	// closing body when function is popped from stack
	defer res.Body.Close()

	// reading the body
	body, err := io.ReadAll(res.Body)
	if err != nil { fmt.Println(err.Error()) }

	// parsing body into nice json
	data, err := gabs.ParseJSON(body)
	if err != nil { fmt.Println(err.Error()) }

	elapsedTime := time.Since(startTime)

	// logging result
	if structure.Method == "POST" {
		if res.StatusCode == 201 {
			fmt.Println(Green + "\n" + structure.Method + " -> " + url + " -> " + strconv.Itoa(res.StatusCode) + " Time: " +
				elapsedTime.Abs().String() + Reset)
		} else {
			fmt.Println(Red + "\n" + structure.Method + " -> " + url + " -> " + strconv.Itoa(res.StatusCode) + " Time: " +
				elapsedTime.Abs().String() + Reset)
		}
	} else {
		if res.StatusCode == 200 {
			fmt.Println(Green + "\n" + structure.Method + " -> " + url + " -> " + strconv.Itoa(res.StatusCode) + " Time: " +
				elapsedTime.Abs().String() + Reset)
		} else {
			fmt.Println(Red + "\n" + structure.Method + " -> " + url + " -> " + strconv.Itoa(res.StatusCode) + " Time: " +
				elapsedTime.Abs().String() + Reset)
		}
	}
	
	// returning response
	return ApiResponse{
		StatusCode: res.StatusCode,
		Body: data,
	}
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
	tokenCheckResponse:= fileContents.Hit(ApiStructure{
		Endpoint: "/me",
	})

	// if request fails with unauthorized, generate new token
	if tokenCheckResponse.StatusCode == 401 {
		fmt.Println(Red + "- Invalid token found..." + Reset)
		GetAndStoreToken(fileContents)
		return 
	}

	fmt.Println(Green + "\nValid token found!!\n" + Reset)
}

func GetAndStoreToken(fileContents* Structure) {

	fmt.Println(Green + "- Generating and storing new token..." + Reset)
	// hitting login api with credentials
	tokenGetResponse := fileContents.Hit(ApiStructure{
		Endpoint: "/login",
		Method: "POST",
		Body: map[string]string{
			"email": fileContents.Credentials.Email,
			"password": fileContents.Credentials.Password,
		},
	})

	// fetching token form the json response from the given structure in json file
	token, _ := tokenGetResponse.Body.Path(fileContents.LoginDetails.TokenLocation).Data().(string)

	// storing token in the file and in app state
	os.WriteFile("token.txt", []byte(token), 0633)
	fileContents.LoginDetails.Token = token
}

func (fileContents* Structure) CallCurrentPipeline() {

	fmt.Println(Blue + "\nCalling All API in current pipeline\n" + Reset)
	for i := range fileContents.PipelineBody {
		res := fileContents.Hit(ApiStructure{
			Endpoint: fileContents.PipelineBody[i].Endpoint,
			Method: fileContents.PipelineBody[i].Method,
			Body: fileContents.PipelineBody[i].Body,
		})

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

			res := fileContents.Hit(ApiStructure{
				Endpoint: data["endpoint"].Data().(string),
				Method: data["method"].Data().(string),
				Body: data["body"].Data(),
			})

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

				res := fileContents.Hit(ApiStructure{
					Endpoint: data["endpoint"].Data().(string),
					Method: data["method"].Data().(string),
					Body: data["body"].Data(),
				})

				fmt.Println(res.Body.StringIndent("", "  "))
			}
		}
	}
}

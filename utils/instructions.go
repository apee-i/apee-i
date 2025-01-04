package utils

import (
	"fmt"
	"io"
	"os"
	"encoding/json"
)

type Credentials struct {
	Development any `json:"development"`
	Staging any `json:"staging"`
	Production any `json:"production"`
}

type Environments struct {
	Development string `json:"development"`
	Staging string `json:"staging"`
	Production string `json:"production"`
}

type LoginDetails struct {
	Route string `json:"route"`
	TokenLocation string `json:"token_location"`
	Token string 
}

type PipelineBody struct {
	Method string `json:"method"`
	Endpoint string `json:"endpoint"`
	Body Credentials `json:"body"`
	Headers any `json:"headers"`
	ExpectedStatusCode int `json:"expectedStatusCode"`
	ExpectedBody any `json:"expectedBody"`
}

type Structure struct {
	BaseUrl Environments `json:"baseUrl"`
	Credentials Credentials `json:"credentials"`
	LoginDetails LoginDetails `json:"loginDetails"`
	PipelineBody []PipelineBody `json:"current_pipeline"`
	CustomPipelines map[string]interface{} `json:"custom_pipelines"`
	ActiveUrl string
	ActiveEnvironment string
}

func (fileContents *Structure) ReadInstructions(filepath string) error {
	
	file, err := os.Open(filepath)
	if err != nil { return fmt.Errorf("Could not open file") }

	defer file.Close()

	fileRawContents, err := io.ReadAll(file)
	if err != nil { return fmt.Errorf("Could not read file contents") }
	
	json.Unmarshal(fileRawContents, &fileContents)
	return nil
}

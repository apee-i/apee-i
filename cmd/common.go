package cmd

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
)

type ApiStructure struct {
	Method             string
	Endpoint           string
	Body               any
	ExpectedStatusCode int
	ExpectedBody       any
	Headers            any
}

type ApiResponse struct {
	StatusCode int
	Body       *gabs.Container
}

// Credentials contains all the login properties
// with respect to the 3 different environments
type Credentials struct {
	Development any `yaml:"development" json:"development"`
	Staging any `yaml:"staging" json:"staging"`
	Production any `yaml:"production" json:"production"`
}

type Environments struct {
	Development string `yaml:"development" json:"development"`
	Staging string `yaml:"staging" json:"staging"`
	Production string `yaml:"production" json:"production"`
}

type LoginDetails struct {
	Route string `yaml:"route" json:"route"`
	TokenLocation string `yaml:"token_location" json:"token_location"`
	Token string 
}

type PipelineBody struct {
	Method string `yaml:"method" json:"method"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Body any `yaml:"body" json:"body"`
	Headers any `yaml:"headers" json:"headers"`
	ExpectedStatusCode int `yaml:"expectedStatusCode" json:"expectedStatusCode"`
	ExpectedBody any `yaml:"expectedBody" json:"expectedBody"`
}

type Structure struct {
	BaseUrl Environments `yaml:"baseUrl" json:"baseUrl"`
	Credentials Credentials `yaml:"credentials" json:"credentials"`
	LoginDetails LoginDetails `yaml:"loginDetails" json:"loginDetails"`
	PipelineBody []PipelineBody `yaml:"current_pipeline" json:"current_pipeline"`
	CustomPipelines map[string]interface{} `yaml:"custom_pipelines" json:"custom_pipelines"`
	ActiveUrl string
	ActiveEnvironment string
}

type FileReaderStrategy interface {
	ReadInstructions(filepath string) (*Structure, error)
	Login(fileContents *Structure)
	CallCurrentPipeline(fileContents *Structure)
	CallCustomPipelines(fileContents *Structure)
	CallSingleCustomPipeline(fileContents *Structure, pipelineKey string)
}

// Context for reading file instructions
type FileReaderContext struct {
	strategy FileReaderStrategy
}

// SetStrategy allows you to change the strategy at runtime
func (c *FileReaderContext) SetStrategy(strategy FileReaderStrategy) {
	c.strategy = strategy
}

// ReadInstructions delegates the reading task to the strategy
func (c *FileReaderContext) ReadInstructions(filepath string) (*Structure, error) {
	if c.strategy == nil {
		return nil, fmt.Errorf("strategy not set")
	}

	structure, err := c.strategy.ReadInstructions(filepath)
	if err != nil { return nil, fmt.Errorf("Could not call strategy") }

	return structure, nil
}

// Login delegates the reading task to the strategy
func (c *FileReaderContext) Login(fileContents *Structure) {
	if c.strategy == nil { fmt.Println("strategy not set"); return }

	c.strategy.Login(fileContents)
}

// Call Current pipeline delegates the reading task to the strategy
func (c *FileReaderContext) CallCurrentPipeline(fileContents *Structure) {
	if c.strategy == nil { fmt.Println("strategy not set"); return }

	c.strategy.CallCurrentPipeline(fileContents)
}

// Call All pipeline delegates the reading task to the strategy
func (c *FileReaderContext) CallCustomPipelines(fileContents *Structure) {
	if c.strategy == nil { fmt.Println("strategy not set"); return }

	c.strategy.CallCustomPipelines(fileContents)
}

// Call All pipeline delegates the reading task to the strategy
func (c *FileReaderContext) CallSingleCustomPipeline(fileContents *Structure, pipelineKey string) {
	if c.strategy == nil { fmt.Println("strategy not set"); return }

	c.strategy.CallSingleCustomPipeline(fileContents, pipelineKey)
}



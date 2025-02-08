package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/utils"
	"github.com/Jeffail/gabs/v2"
)

// Strategy is used to make selection for the strategy
type Strategy struct{}

// Hit acts as an HTTP client and hits a rest based request
func (s *Strategy) Hit(fileContents *cmd.Structure, structure cmd.PipelineBody) (cmd.APIResponse, error) {

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

	// adding custom headers from the user
	if structure.Headers != nil {
		for key, value := range structure.Headers {
			req.Header.Set(key, value.(string))
		}
	}
	// adding appropriate headers
	req.Header.Set("Authorization", "bearer "+fileContents.LoginDetails.Token)

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

package inline

import (
	"fmt"
	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/cmd/protocols"
	myHttp "github.com/IbraheemHaseeb7/apee-i/cmd/protocols/http"
	"github.com/IbraheemHaseeb7/apee-i/utils"
	"net/url"
)

// Options contains the properties that define an inline request
type Options struct {
	Method             string
	URL                string
	Body               string
	Headers            map[string]string
	Protocol           string
	Timeout            int
	Token              string
	ExpectedStatusCode int
}

// RunInline function executes an inline request
func RunInline(opts Options) {
	parsedURL, err := url.Parse(opts.URL)

	if err != nil {
		fmt.Println(utils.Red + "Error parsing the URL: " + err.Error() + utils.Reset)
	}

	baseURL := parsedURL.Scheme + "://" + parsedURL.Host
	endpoint := parsedURL.Path

	pipeline := *cmd.NewPipelineBody(&cmd.PipelineBody{
		Method:             opts.Method,
		Body:               opts.Body,
		Endpoint:           endpoint,
		Headers:            convertHeaders(opts.Headers),
		BaseURL:            baseURL,
		Protocol:           opts.Protocol,
		Timeout:            opts.Timeout,
		ExpectedStatusCode: opts.ExpectedStatusCode,
	})

	protocolContext := &protocols.ProtocolContext{}

	if opts.Protocol == "HTTP" {
		protocolContext.SetStrategy(&myHttp.Strategy{})
	} else {
		fmt.Println(utils.Red + "Invalid Protocol specified. Only HTTP is supported at this time" + opts.Protocol)
		return
	}

	if opts.Token != "" {
		if pipeline.Headers == nil {
			pipeline.Headers = make(map[string]interface{})
		}
		pipeline.Headers["Authorization"] = "Bearer " + opts.Token
	}

	dummyConfig := &cmd.Structure{
		ActiveURL: baseURL,
		LoginDetails: cmd.LoginDetails{
			Token: opts.Token,
		},
	}

	res, err := protocolContext.Hit(dummyConfig, pipeline)

	if err != nil {
		fmt.Println(utils.Red + "Error executing inline call: " + err.Error() + utils.Reset)
	}

	fmt.Println("Response:")
	fmt.Println(res.Body.StringIndent("", " "))
}

func convertHeaders(input map[string]string) map[string]any {
	headers := make(map[string]any)
	for k, v := range input {
		headers[k] = v
	}

	return headers
}

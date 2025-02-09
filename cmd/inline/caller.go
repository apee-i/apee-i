package inline

import (
	"fmt"
	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/IbraheemHaseeb7/apee-i/cmd/protocols"
	myHttp "github.com/IbraheemHaseeb7/apee-i/cmd/protocols/http"
	"github.com/IbraheemHaseeb7/apee-i/utils"
)

type Options struct {
	Method             string
	Body               string
	Endpoint           string
	Headers            map[string]string
	BaseURL            string
	Protocol           string
	Timeout            int
	Token              string
	ExpectedStatusCode int
}

func RunInline(opts Options) {
	pipeline := *cmd.NewPipelineBody(&cmd.PipelineBody{
		Method:             opts.Method,
		Body:               opts.Body,
		Endpoint:           opts.Endpoint,
		Headers:            convertHeaders(opts.Headers),
		BaseURL:            opts.BaseURL,
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
		ActiveURL: opts.BaseURL,
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

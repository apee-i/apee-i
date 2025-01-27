package protocols

import (
	"fmt"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
)

// ProtocolStrategy allows the dynamic behaviour in changing protocols
type ProtocolStrategy interface {
	Hit(fileContents *cmd.Structure, structure cmd.PipelineBody) (cmd.APIResponse, error)
}

// ProtocolContext is used to change function calls based on selected strategy
type ProtocolContext struct {
	strategy ProtocolStrategy
}

// SetStrategy is used to set protocol strategy
func (p *ProtocolContext) SetStrategy(strategy ProtocolStrategy) {
	p.strategy = strategy
}

// Hit method is the final endpoint of any request that interacts with the host
func (p *ProtocolContext) Hit(fileContents *cmd.Structure, structure cmd.PipelineBody) (cmd.APIResponse, error) {
	if p.strategy == nil {
		return cmd.APIResponse{}, fmt.Errorf("strategy not set")
	}

	return p.strategy.Hit(fileContents, structure)
}

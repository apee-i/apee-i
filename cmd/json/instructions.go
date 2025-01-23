package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/apee-i/apee-i/cmd"
)

// Reader purpose is just to allow the programmer to make selection for JSON
type Reader struct{}

// ReadInstructions decodes and stores all the instructions from the configuration
// file in the state of the program
func (r *Reader) ReadInstructions(filepath string) (*cmd.Structure, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return &cmd.Structure{}, fmt.Errorf("Could not open file")
	}

	defer file.Close()

	fileRawContents, err := io.ReadAll(file)
	if err != nil {
		return &cmd.Structure{}, fmt.Errorf("Could not read file contents")
	}

	fileContents := new(cmd.Structure)
	json.Unmarshal(fileRawContents, &fileContents)

	return fileContents, nil
}

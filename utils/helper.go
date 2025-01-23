package utils

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/IbraheemHaseeb7/apee-i/cmd"
	"github.com/jedib0t/go-pretty/v6/table"
)

// ResponseLogger is used to print out the API response in a nice green or red colored table for easier read
func ResponseLogger(structure cmd.PipelineBody , res *http.Response, url string, elapsedTime time.Duration) {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Method", "URL", "Got Status Code", "Expected Status Code", "Time Lapsed"})
	t.AppendSeparator()
	expectedStatusCode := "Not Given"
	if structure.ExpectedStatusCode == 0 {
		if structure.Method == "POST" {

			if res.StatusCode == 201 || res.StatusCode == 200 {
				t.SetStyle(table.StyleColoredBlackOnGreenWhite)
			} else {
				t.SetStyle(table.StyleColoredBlackOnRedWhite)
			}
		} else {
			if res.StatusCode == 200 {
				t.SetStyle(table.StyleColoredBlackOnGreenWhite)
			} else {
				t.SetStyle(table.StyleColoredBlackOnRedWhite)
			}
		}
	} else {
		expectedStatusCode = strconv.Itoa(structure.ExpectedStatusCode)
		if res.StatusCode == structure.ExpectedStatusCode {
			t.SetStyle(table.StyleColoredBlackOnGreenWhite)
		} else {
			t.SetStyle(table.StyleColoredBlackOnRedWhite)
		}
	}

	t.AppendRow(table.Row{structure.Method, url, strconv.Itoa(res.StatusCode), expectedStatusCode, elapsedTime.Abs().String()})
	t.Render()

}

// func ValidateExpectedBody() {
//
// }

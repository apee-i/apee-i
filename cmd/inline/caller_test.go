package inline

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	reader, writer, err := os.Pipe()

	if err != nil {
		panic(err)
	}

	// redirect standard output to the writer
	os.Stdout = writer

	f()

	writer.Close()

	var buf bytes.Buffer
	io.Copy(&buf, reader)

	// restore the standard output
	os.Stdout = old

	return buf.String()
}

func TestRunInline_InvalidURL(t *testing.T) {
	opts := Options{
		Method:   "GET",
		URL:      "://www.google.com",
		Protocol: "HTTP",
	}

	output := captureOutput(func() {
		RunInline(opts)
	})

	expected := "Error parsing the URL:"

	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, but got %q", expected, output)
	}
}

func TestRunInline_InvalidProtocol(t *testing.T) {
	opts := Options{
		Method:   "GET",
		URL:      "https://www.google.com",
		Protocol: "SMTP",
	}

	output := captureOutput(func() {
		RunInline(opts)
	})

	expected := "Invalid Protocol specified. Only HTTP is supported at this time"

	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, but got %q", expected, output)
	}
}

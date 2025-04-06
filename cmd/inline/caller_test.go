package inline

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestRunInline_ConvertHeaders(t *testing.T) {
	input := map[string]string{
		"Content-Type": "application/json",
	}

	headers := convertHeaders(input)

	if headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type to be application/json, but got %q", headers["Content-Type"])
	}
}

func TestRunInline_Integration(t *testing.T) {
	// set up a dummy http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected method GET, but got %s", r.Method)
		}

		if authHeader := r.Header.Get("Authorization"); authHeader != "bearer dummy-token" {
			t.Errorf("Expected Authorization header to be %q but got %q", "bearer dummy-token", authHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer ts.Close()

	parsedURL, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	opts := Options{
		Method:   "GET",
		URL:      parsedURL.String() + "/api",
		Protocol: "HTTP",
		Token:    "dummy-token",
	}

	output := captureOutput(func() {
		RunInline(opts)
	})

	if !strings.Contains(output, "Response:") {
		t.Errorf("Expected output to contain %q, got %s", `Response:`, output)
	}

	if !strings.Contains(output, "success") {
		t.Errorf("Expected output to coontain %q, got %s", "success", output)
	}

}

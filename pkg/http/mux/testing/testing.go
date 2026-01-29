package testing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type Args struct {
	Method                string
	Path                  string
	Headers               [][2]string
	Body                  []byte
	ExpectedStatusCode    int
	ExpectedHeaders       [][2]string
	ExpectedBody          []byte
	ExpectedProblemDetail *problem_detail.Detail
	ExpectedClientDoError error
}

func TestArgs(t *testing.T, args *Args, serverUrl string) {
	t.Helper()

	if args == nil {
		t.Fatalf("args is nil")
	}

	if serverUrl == "" {
		t.Fatalf("server url is empty")
	}

	var requestBody io.Reader
	if testCaseBody := args.Body; len(testCaseBody) != 0 {
		requestBody = bytes.NewReader(testCaseBody)
	}

	request, err := http.NewRequest(args.Method, serverUrl+args.Path, requestBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	for _, header := range args.Headers {
		request.Header.Set(header[0], header[1])
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if args.ExpectedClientDoError != nil {
			if errors.Is(err, args.ExpectedClientDoError) {
				return
			}
			t.Fatalf("http client do: %v", err)
		}
	}
	if response == nil {
		t.Fatalf("http client do returned nil response")
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("io read all response body: %v", err)
	}

	if response.StatusCode != args.ExpectedStatusCode {
		t.Errorf("got status code %d, expected %d", response.StatusCode, args.ExpectedStatusCode)
	}

	if expectedHeaders := args.ExpectedHeaders; len(expectedHeaders) != 0 {
		responseHeader := response.Header
		for _, header := range expectedHeaders {
			headerValue := responseHeader.Get(header[0])
			if headerValue != header[1] {
				t.Errorf("got %q, expected header %q to be %q", headerValue, header[0], header[1])
			}
		}
	}

	if expectedProblemDetail := args.ExpectedProblemDetail; expectedProblemDetail != nil {
		var problemDetail *problem_detail.Detail
		if err := json.Unmarshal(responseBody, &problemDetail); err != nil {
			t.Fatalf("json unmarshal response body: %v", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreFields(problem_detail.Detail{}, "Type"),
			cmpopts.IgnoreFields(problem_detail.Detail{}, "Instance"),
			cmpopts.EquateEmpty(),
		}

		expectedStatusCode := args.ExpectedStatusCode
		expectedProblemDetail.Title = http.StatusText(expectedStatusCode)
		expectedProblemDetail.Status = expectedStatusCode

		if diff := cmp.Diff(expectedProblemDetail, problemDetail, opts...); diff != "" {
			t.Errorf("problem detail mismatch (-expected +got):\n%s", diff)
		}
	} else if !bytes.Equal(responseBody, args.ExpectedBody) {
		t.Errorf("got response body %q, expected response body %q", responseBody, args.ExpectedBody)
	}
}

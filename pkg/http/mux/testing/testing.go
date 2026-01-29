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

type Case struct {
	Name                  string
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

func TestMuxCase(testCase *Case, serverUrl string, t *testing.T) {
	t.Helper()

	var requestBody io.Reader
	if testCaseBody := testCase.Body; len(testCaseBody) != 0 {
		requestBody = bytes.NewReader(testCaseBody)
	}

	request, err := http.NewRequest(testCase.Method, serverUrl+testCase.Path, requestBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	for _, header := range testCase.Headers {
		request.Header.Set(header[0], header[1])
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if testCase.ExpectedClientDoError != nil {
			if errors.Is(err, testCase.ExpectedClientDoError) {
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

	if response.StatusCode != testCase.ExpectedStatusCode {
		t.Errorf("got status code %d, expected %d", response.StatusCode, testCase.ExpectedStatusCode)
	}

	if expectedHeaders := testCase.ExpectedHeaders; len(expectedHeaders) != 0 {
		responseHeader := response.Header
		for _, header := range expectedHeaders {
			headerValue := responseHeader.Get(header[0])
			if headerValue != header[1] {
				t.Errorf("got %q, expected header %q to be %q", headerValue, header[0], header[1])
			}
		}
	}

	if expectedProblemDetail := testCase.ExpectedProblemDetail; expectedProblemDetail != nil {
		var problemDetail *problem_detail.Detail
		if err := json.Unmarshal(responseBody, &problemDetail); err != nil {
			t.Fatalf("json unmarshal response body: %v", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreFields(problem_detail.Detail{}, "Type"),
			cmpopts.IgnoreFields(problem_detail.Detail{}, "Instance"),
			cmpopts.EquateEmpty(),
		}

		expectedStatusCode := testCase.ExpectedStatusCode
		expectedProblemDetail.Title = http.StatusText(expectedStatusCode)
		expectedProblemDetail.Status = expectedStatusCode

		if diff := cmp.Diff(expectedProblemDetail, problemDetail, opts...); diff != "" {
			t.Errorf("problem detail mismatch (-expected +got):\n%s", diff)
		}
	} else if !bytes.Equal(responseBody, testCase.ExpectedBody) {
		t.Errorf("got response body %q, expected response body %q", responseBody, testCase.ExpectedBody)
	}
}

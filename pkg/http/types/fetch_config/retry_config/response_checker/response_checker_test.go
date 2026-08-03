package response_checker

import (
	"errors"
	"net/http"
	"testing"
)

var errBoom = errors.New("boom")

func TestResponseCheckerFunctionCheck(t *testing.T) {
	t.Parallel()

	var gotResponse *http.Response
	var gotErr error

	sentinelResponse := &http.Response{StatusCode: http.StatusTeapot}
	sentinelErr := errBoom

	checker := ResponseCheckerFunction(func(response *http.Response, err error) bool {
		gotResponse = response
		gotErr = err
		return response != nil
	})

	if !checker.Check(sentinelResponse, sentinelErr) {
		t.Error("Check() = false, want true")
	}
	if gotResponse != sentinelResponse {
		t.Errorf("response passed = %p, want %p", gotResponse, sentinelResponse)
	}
	if !errors.Is(gotErr, sentinelErr) {
		t.Errorf("err passed = %v, want %v", gotErr, sentinelErr)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		fn       func(*http.Response, error) bool
		response *http.Response
		err      error
		want     bool
	}{
		{
			name:     "always true",
			fn:       func(*http.Response, error) bool { return true },
			response: &http.Response{},
			want:     true,
		},
		{
			name: "based on status code",
			fn: func(response *http.Response, _ error) bool {
				return response != nil && response.StatusCode >= 500
			},
			response: &http.Response{StatusCode: http.StatusServiceUnavailable},
			want:     true,
		},
		{
			name: "based on status code below threshold",
			fn: func(response *http.Response, _ error) bool {
				return response != nil && response.StatusCode >= 500
			},
			response: &http.Response{StatusCode: http.StatusOK},
			want:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			checker := New(testCase.fn)
			if checker == nil {
				t.Fatal("New() returned nil")
			}
			if got := checker.Check(testCase.response, testCase.err); got != testCase.want {
				t.Errorf("Check() = %v, want %v", got, testCase.want)
			}
		})
	}
}

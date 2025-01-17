package retry_after

import (
	"testing"
	"time"
)

func TestRetryAfterBadInput(t *testing.T) {
	data := []byte("so bad")

	retryAfter, err := ParseRetryAfter(data)
	if err != nil {
		t.Fatalf("an error occurred when parsing the data: %v", err)
	}
	if retryAfter != nil {
		t.Fatal("retry after parsing succeeded with bad input")
	}
}

func TestRetryAfterHttpDate(t *testing.T) {
	data := []byte("Fri, 31 Dec 1999 23:59:59 GMT")

	retryAfter, err := ParseRetryAfter(data)
	if err != nil {
		t.Fatalf("an error occurred when parsing the data: %v", err)
	}
	if retryAfter == nil {
		t.Fatal("retry after is nil")
	}

	httpDate, ok := retryAfter.WaitTime.(time.Time)
	if !ok {
		t.Fatal("the wait time could not be converted to a time.Time")
	}

	if httpDate.Unix() != 946684799 {
		t.Fatal("the http date was not parsed as the correct time")
	}
}

func TestRetryAfterDelay(t *testing.T) {
	data := []byte("120")

	retryAfter, err := ParseRetryAfter(data)
	if err != nil {
		t.Fatalf("an error occurred when parsing the data: %v", err)
	}
	if retryAfter == nil {
		t.Fatal("retry after is nil")
	}

	delaySeconds, ok := retryAfter.WaitTime.(time.Duration)
	if !ok {
		t.Fatal("the wait time could not be converted to int64")
	}

	expectedDuration := 120 * time.Second

	if delaySeconds != expectedDuration {
		t.Fatal("the delay seconds was not parsed as the correct duration")
	}
}

package debug

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/fastly/go-fastly/v14/fastly"
)

// TestParseLoggingError validates we're correctly decoding individual logging error JSON.
func TestParseLoggingError(t *testing.T) {
	data := []byte(`{"sequence_number":1,"error_time_us":1601645172164,"stream":"logging_error","message":"Failed to send log","endpoint":"my-s3-endpoint","details":"connection refused"}`)

	var got fastly.LoggingEndpointError
	err := json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("error parsing response data: %v", err)
	}

	want := fastly.LoggingEndpointError{
		SequenceNumber: 1,
		Timestamp:      1601645172164,
		Stream:         "logging_error",
		Message:        "Failed to send log",
		Endpoint:       "my-s3-endpoint",
		Details:        "connection refused",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("JSON unmarshal mismatch (-want +got):\n%s", diff)
	}
}

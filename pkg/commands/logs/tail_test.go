package logs

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const responseFile = "testdata/response.json"

// TestAdjustTimes tests that the from and to times are adjusted accordingly
// based on the searchPadding.
func TestAdjustTimes(t *testing.T) {
	dur, _ := time.ParseDuration("10s")
	for i, test := range []struct {
		in  cfg
		exp cfg
	}{
		{
			in:  cfg{from: 1601480668, to: 1601480768, searchPadding: dur},
			exp: cfg{from: 1601480658, to: 1601480778, searchPadding: dur},
		},
		{
			in:  cfg{from: 1601480668, to: 1601480768},
			exp: cfg{from: 1601480668, to: 1601480768},
		},
		{
			in:  cfg{searchPadding: dur},
			exp: cfg{searchPadding: dur},
		},
	} {
		c := TailCommand{cfg: test.in}
		c.adjustTimes()
		if equal := reflect.DeepEqual(test.exp, c.cfg); !equal {
			t.Errorf("#%d: adjustTimes mismatch: got: %#+v  want: %#+v", i, c.cfg, test.exp)
		}
	}
}

// TestSplitByReqID tests that logs are properly grouped and sorted
// by their RequestID and SequenceNum.
func TestSplitByReqID(t *testing.T) {
	full := []Log{
		{SequenceNum: 1, RequestID: "41f82900"},
		{SequenceNum: 2, RequestID: "41f82900"},
		{SequenceNum: 3, RequestID: "2bef4613"},
		{SequenceNum: 4, RequestID: "2bef4613"},
		{SequenceNum: 5, RequestID: "41f82900"},
		{SequenceNum: 6, RequestID: "41f82900"},
		{SequenceNum: 6, RequestID: "2bef4613"},
		{SequenceNum: 1, RequestID: "2bef4613"},
		{SequenceNum: 5, RequestID: "2bef4613"},
		{SequenceNum: 2, RequestID: "2bef4613"},
		{SequenceNum: 3, RequestID: "41f82900"},
		{SequenceNum: 4, RequestID: "41f82900"},
	}

	expfull := map[string][]Log{
		"41f82900": {
			{SequenceNum: 1, RequestID: "41f82900"},
			{SequenceNum: 2, RequestID: "41f82900"},
			{SequenceNum: 5, RequestID: "41f82900"},
			{SequenceNum: 6, RequestID: "41f82900"},
			{SequenceNum: 3, RequestID: "41f82900"},
			{SequenceNum: 4, RequestID: "41f82900"},
		},
		"2bef4613": {
			{SequenceNum: 3, RequestID: "2bef4613"},
			{SequenceNum: 4, RequestID: "2bef4613"},
			{SequenceNum: 6, RequestID: "2bef4613"},
			{SequenceNum: 1, RequestID: "2bef4613"},
			{SequenceNum: 5, RequestID: "2bef4613"},
			{SequenceNum: 2, RequestID: "2bef4613"},
		},
	}

	single := []Log{
		{SequenceNum: 1, RequestID: "41f82900"},
	}

	expsingle := map[string][]Log{
		"41f82900": {{SequenceNum: 1, RequestID: "41f82900"}},
	}

	for i, test := range []struct {
		in   []Log
		want map[string][]Log
	}{
		{in: full, want: expfull},
		{in: single, want: expsingle},
		{in: []Log{}, want: make(map[string][]Log)},
	} {
		got := splitByReqID(test.in)
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("#%d: splitByReqID mismatch (-want +got):\n%s", i, diff)
		}
	}
}

// TestParseResponseData validates we're correctly decoding a batch JSON log
// response into a logs.Batch type.
func TestParseResponseData(t *testing.T) {
	data, err := os.ReadFile(responseFile)
	if err != nil {
		t.Fatalf("cannot read from file: %v", err)
	}

	got, err := parseResponseData(data)
	if err != nil {
		t.Fatalf("error parsing response data: %v", err)
	}

	want := Batch{
		ID: "MC0x",
		Logs: []Log{
			{SequenceNum: 1, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1"},
			{SequenceNum: 2, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "2 2 2 2 2 2 2 2 2 2 2 2 2 2 2 2 2 2 2 2 2"},
			{SequenceNum: 3, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "3 3 3 3 3 3 3 3 3 3 3 3 3 3 3 3 3 3 3 3 3"},
			{SequenceNum: 4, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "4 4 4 4 4 4 4 4 4 4 4 4 4 4 4 4 4 4 4 4 4"},
			{SequenceNum: 5, RequestStart: 1601645172164667, Stream: "stderr", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "5 5 5 5 5 5 5 5 5 5 5 5 5 5 5 5 5 5 5 5 5"},
			{SequenceNum: 6, RequestStart: 1601645172164667, Stream: "stderr", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "6 6 6 6 6 6 6 6 6 6 6 6 6 6 6 6 6 6 6 6 6"},
			{SequenceNum: 7, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "7 7 7 7 7 7 7 7 7 7 7 7 7 7 7 7 7 7 7 7 7"},
			{SequenceNum: 8, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8 8"},
			{SequenceNum: 9, RequestStart: 1601645172164667, Stream: "stderr", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9 9"},
			{SequenceNum: 10, RequestStart: 1601645172164667, Stream: "stdout", RequestID: "44a1eedd-5831-49fe-b094-7435908ba1fb", Message: "10 10 10 10 10 10 10 10 10 10 10 10 10 10"},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResponseData mismatch (-want +got):\n%s", diff)
	}
}

// TestFilterStream tests that a passed in stream will filter out
// unwanted output.
func TestFilterStream(t *testing.T) {
	for i, test := range []struct {
		stream string
		logs   []Log
		explen int
	}{
		{
			stream: "stdout",
			logs: []Log{
				{Stream: "stdout"},
				{Stream: "stdout"},
				{Stream: "stderr"},
				{Stream: "stdout"},
				{Stream: "stdout"},
				{Stream: "stderr"},
				{Stream: "stdout"},
				{Stream: "stderr"},
			},
			explen: 5,
		},
		{
			stream: "stderr",
			logs: []Log{
				{Stream: "stdout"},
				{Stream: "stdout"},
				{Stream: "stderr"},
				{Stream: "stdout"},
				{Stream: "stdout"},
				{Stream: "stderr"},
				{Stream: "stdout"},
				{Stream: "stderr"},
			},
			explen: 3,
		},
		{
			logs: []Log{
				{Stream: "stdout"},
				{Stream: "stdout"},
				{Stream: "stderr"},
				{Stream: "stderr"},
				{Stream: "stdout"},
				{Stream: "stderr"},
				{Stream: "stdout"},
				{Stream: "stderr"},
			},
			explen: 8,
		},
	} {
		out := filterStream(test.stream, test.logs)
		if len(out) != test.explen {
			t.Errorf("#%d: exp: %d != got: %d", i, test.explen, len(out))
		}
	}
}

// TestGetLinks tests that we can parse next and prev links from a Link HTTP
// header.
func TestGetLinks(t *testing.T) {
	rawNexPrev := `</service/sid/log_stream/managed/instance_output%3Ffrom=1601412640>; rel="next", </service/sid/log_stream/managed/instance_output%3Ffrom=1601412620>; rel="prev"`
	head := make(http.Header)
	head.Set("Link", rawNexPrev)

	prev, next := getLinks(head)
	prevexp := "/service/sid/log_stream/managed/instance_output%3Ffrom=1601412620"
	if prev != prevexp {
		t.Errorf("prev header exp: %s != got: %s", prevexp, prev)
	}

	nextexp := "/service/sid/log_stream/managed/instance_output%3Ffrom=1601412640"
	if next != nextexp {
		t.Errorf("next header exp: %s != got: %s", nextexp, next)
	}

	pTime, err := getTimeFromLink(prev)
	if err != nil {
		t.Fatalf("unexpected error parsing prev link: %s", err)
	}
	exp := int64(1601412620)
	if pTime != exp {
		t.Errorf("prev time exp: %v != got: %v", exp, pTime)
	}

	nTime, err := getTimeFromLink(next)
	if err != nil {
		t.Fatalf("unexpected error parsing next link: %s", err)
	}
	exp = int64(1601412640)
	if nTime != exp {
		t.Errorf("next time exp: %v != got: %v", exp, nTime)
	}
}

// TestSplitOnIdx tests both findIdxBySeq() and the split functionality
// that is used when a timer expires in the outputLoop() function. Indexes
// and splitting (especially at slice boundaries) are particularly error prone.
func TestSplitOnIdx(t *testing.T) {
	for i, test := range []struct {
		seq      int
		logs     []Log
		expleft  int
		expright int
	}{
		{
			seq: 4,
			logs: []Log{
				{SequenceNum: 0},
				{SequenceNum: 1},
				{SequenceNum: 2},
				{SequenceNum: 3},
				{SequenceNum: 4},
				{SequenceNum: 5},
				{SequenceNum: 6},
				{SequenceNum: 7},
			},
			expleft:  5,
			expright: 3,
		},
		{
			seq: 1,
			logs: []Log{
				{SequenceNum: 0},
				{SequenceNum: 1},
				{SequenceNum: 2},
				{SequenceNum: 3},
			},
			expleft:  2,
			expright: 2,
		},
		{
			seq: 4,
			logs: []Log{
				{SequenceNum: 1},
				{SequenceNum: 2},
				{SequenceNum: 3},
				{SequenceNum: 4},
			},
			expleft:  4,
			expright: 0,
		},
		{
			seq: 6,
			logs: []Log{
				{SequenceNum: 1},
				{SequenceNum: 3},
				{SequenceNum: 5},
			},
			expleft:  3,
			expright: 0,
		},
	} {
		idx := findIdxBySeq(test.logs, test.seq)
		left, right := test.logs[:idx], test.logs[idx:]
		if len(left) != test.expleft {
			t.Errorf("#%d: exp: %d != got: %d", i, test.expleft, len(left))
		}

		if len(right) != test.expright {
			t.Errorf("#%d: exp: %d != got: %d", i, test.expright, len(right))
		}

	}
}

// TestHighSequence tests that we correctly get the highest
// SequenceNum in a slice of logs.
func TestHighSequence(t *testing.T) {
	for i, test := range []struct {
		logs []Log
		exp  int
	}{
		{
			logs: []Log{
				{SequenceNum: 0},
				{SequenceNum: 1},
				{SequenceNum: 2},
			},
			exp: 2,
		},
		{
			logs: []Log{
				{SequenceNum: 2},
				{SequenceNum: 1},
				{SequenceNum: 0},
			},
			exp: 2,
		},
		{
			logs: []Log{
				{SequenceNum: 1},
				{SequenceNum: 1},
			},
			exp: 1,
		},
		{
			logs: []Log{},
			exp:  0,
		},
	} {
		if got := highSequence(test.logs); got != test.exp {
			t.Errorf("#%d: exp: %d != got: %d", i, test.exp, got)
		}
	}
}

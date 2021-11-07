package awsping

import (
	"bytes"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	input := Measure(1 * time.Second)
	want := "1000.0 ms"
	got := input.toStr()
	if got != want {
		t.Errorf("Duration was incorrect, got: %s, want: %s.", got, want)
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		n   int
		res string
	}{
		{1, ""},
		{5, ""},
		{10, ""},
	}

	for idx, test := range tests {
		test.res = mkRandoString(test.n)
		if len(test.res) != test.n {
			t.Errorf("Try %d: n=%d, got: %s (len=%d), want: %d.",
				idx, test.n, test.res, len(test.res), test.n)
		}
	}
}

func TestOutputShowOnlyRegions(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(ShowOnlyRegions, 0)
	lo.w = &b

	regions := GetRegions("sqs")[:2]
	lo.Show(&regions)

	got := b.String()
	want := "us-east-1       US-East (N. Virginia)\n" +
		"us-east-2       US-East (Ohio)      \n"
	if got != want {
		t.Errorf("Show: got=%q\nwant=%q", got, want)
	}
}

func TestOutputShow0(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(0, 0)
	lo.w = &b

	regions := GetRegions("sqs")[:2]
	regions[0].Latencies = []Measure{Measure(15 * time.Millisecond)}
	regions[1].Latencies = []Measure{Measure(25 * time.Millisecond)}

	lo.Show(&regions)

	want := "US-East (N. Virginia)                  15.0 ms\n" +
		"US-East (Ohio)                         25.0 ms\n"
	got := b.String()
	if got != want {
		t.Errorf("Show0 failed:\ngot=%q\nwant=%q", got, want)
	}
}

func TestOutputShow1(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(1, 0)
	lo.w = &b

	regions := GetRegions("sqs")[:2]
	regions[0].Latencies = []Measure{Measure(15 * time.Millisecond)}
	regions[1].Latencies = []Measure{Measure(25 * time.Millisecond)}

	lo.Show(&regions)

	got := b.String()
	want := "      Code            Region                                      Latency\n" +
		"    0 us-east-1       US-East (N. Virginia)                       15.0 ms\n" +
		"    1 us-east-2       US-East (Ohio)                              25.0 ms\n"
	if got != want {
		t.Errorf("Show1 failed:\ngot=%q\nwant=%q", got, want)
	}
}

func TestOutputShow2(t *testing.T) {
	var b bytes.Buffer

	lo := NewOutput(2, 2)
	lo.w = &b

	regions := GetRegions("sqs")[:2]
	regions[0].Latencies = []Measure{Measure(15 * time.Millisecond), Measure(17 * time.Millisecond)}
	regions[1].Latencies = []Measure{Measure(25 * time.Millisecond), Measure(26 * time.Millisecond)}

	lo.Show(&regions)

	got := b.String()
	want := "      Code            Region                             Try #1          Try #2     Avg Latency\n" +
		"    0 us-east-1       US-East (N. Virginia)             15.0 ms         17.0 ms         16.0 ms\n" +
		"    1 us-east-2       US-East (Ohio)                    25.0 ms         26.0 ms         25.5 ms\n"
	if got != want {
		t.Errorf("Show2 failed:\ngot=%q\nwant=%q", got, want)
	}
}

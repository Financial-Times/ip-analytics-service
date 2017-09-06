package hooks

import (
	"testing"
	"time"
)

func TestFormatTimestamp(t *testing.T) {
	tStamp := "2017-09-05T09:40:20.845+0000"
	expected := "2017-09-05T09:40:20.845Z"
	if f := formatTimestamp(tStamp); f != expected {
		t.Errorf("Expected formatted timestamp to be %s but received %s", expected, f)
	}
}

func TestDefaultTimestamp(t *testing.T) {
	tStamp := "invalid"
	f := formatTimestamp(tStamp)
	if _, err := time.Parse(TimeFormat, f); err != nil {
		t.Errorf("Expected timestamp in format %s but got err: %s", TimeFormat, err)
	}
}

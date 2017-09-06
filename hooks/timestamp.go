package hooks

import (
	"strings"
	"time"
)

// TimeFormat specifies the custom layout for event timestamps
const TimeFormat = "2006-01-02T15:04:05.999Z"

func formatTimestamp(t string) string {
	r := strings.NewReplacer("+0000", "Z", "-0000", "Z")
	rfc := r.Replace(t)
	tStamp, err := time.Parse(TimeFormat, rfc)
	if err != nil {
		return time.Now().UTC().Format(TimeFormat)
	}
	return tStamp.UTC().Format("2006-01-02T15:04:05.999Z")
}

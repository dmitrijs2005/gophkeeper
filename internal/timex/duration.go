package timex

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration wraps time.Duration to provide custom JSON unmarshalling.
//
// It supports both string and numeric values in JSON:
//
//   - String values are parsed using time.ParseDuration, e.g. "1s", "500ms", "2h45m".
//   - Numeric values are interpreted as nanoseconds (float64 in JSON).
//
// Example JSON:
//
//	{
//	    "interval": "1s",
//	    "timeout": 5000000000
//	}
//
// In the above, "interval" will be parsed as 1 second, and "timeout" as 5 seconds.
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements the json.Unmarshaler interface for Duration.
//
// It accepts either a string (parsed via time.ParseDuration) or a number
// (interpreted as a nanosecond count). Any other JSON type will cause an error.
//
// Example:
//
//	var d Duration
//	_ = json.Unmarshal([]byte(`"150ms"`), &d)
//	fmt.Println(d) // 150ms
func (duration *Duration) UnmarshalJSON(b []byte) error {
	var unmarshalledJson interface{}

	err := json.Unmarshal(b, &unmarshalledJson)
	if err != nil {
		return err
	}

	switch value := unmarshalledJson.(type) {
	case float64:
		duration.Duration = time.Duration(value)
	case string:
		duration.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration: %#v", unmarshalledJson)
	}

	return nil
}

package timex

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration_UnmarshalJSON_String(t *testing.T) {
	tests := []struct {
		name    string
		jsonIn  string
		want    time.Duration
		wantErr bool
	}{
		{"seconds", `"1s"`, time.Second, false},
		{"milliseconds", `"150ms"`, 150 * time.Millisecond, false},
		{"minutes+seconds", `"1m2s"`, time.Minute + 2*time.Second, false},
		{"negative", `"-250ms"`, -250 * time.Millisecond, false},
		{"zero", `"0s"`, 0, false},
		{"invalid-unit", `"10q"`, 0, true},
		{"not-a-duration", `"abc"`, 0, true},
		{"empty-string", `""`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tt.jsonIn), &d)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && d.Duration != tt.want {
				t.Fatalf("got %v, want %v", d.Duration, tt.want)
			}
		})
	}
}

func TestDuration_UnmarshalJSON_Number(t *testing.T) {
	tests := []struct {
		name    string
		jsonIn  string
		want    time.Duration
		wantErr bool
	}{
		{"nanoseconds-plain", `123`, 123 * time.Nanosecond, false},
		{"one-second-exp", `1e9`, time.Second, false},
		{"five-seconds", `5000000000`, 5 * time.Second, false},
		{"zero", `0`, 0, false},
		// JSON allows only finite numbers, but include a big value sanity check
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tt.jsonIn), &d)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && d.Duration != tt.want {
				t.Fatalf("got %v, want %v", d.Duration, tt.want)
			}
		})
	}
}

func TestDuration_UnmarshalJSON_UnsupportedTypes(t *testing.T) {
	tests := []string{
		`true`,
		`false`,
		`null`,
		`{}`,
		`[]`,
	}

	for _, js := range tests {
		t.Run(js, func(t *testing.T) {
			var d Duration
			if err := json.Unmarshal([]byte(js), &d); err == nil {
				t.Fatalf("expected error for input %s, got none (value=%v)", js, d)
			}
		})
	}
}

func TestDuration_UnmarshalJSON_InStruct(t *testing.T) {
	type cfg struct {
		Interval Duration `json:"interval"`
		Timeout  Duration `json:"timeout"`
	}

	js := []byte(`{"interval":"1s","timeout":5000000000}`)
	var c cfg
	if err := json.Unmarshal(js, &c); err != nil {
		t.Fatalf("unmarshal struct: %v", err)
	}

	if c.Interval.Duration != time.Second {
		t.Fatalf("interval got %v, want %v", c.Interval.Duration, time.Second)
	}
	if c.Timeout.Duration != 5*time.Second {
		t.Fatalf("timeout got %v, want %v", c.Timeout.Duration, 5*time.Second)
	}
}

func TestDuration_RoundTrip_StringNumber(t *testing.T) {
	// Ensures numbers-as-ns and strings parse identically for same duration.
	var a, b Duration
	if err := json.Unmarshal([]byte(`"1s"`), &a); err != nil {
		t.Fatalf("unmarshal string: %v", err)
	}
	if err := json.Unmarshal([]byte(`1000000000`), &b); err != nil {
		t.Fatalf("unmarshal number: %v", err)
	}
	if a.Duration != b.Duration || a.Duration != time.Second {
		t.Fatalf("round-trip mismatch: a=%v b=%v", a.Duration, b.Duration)
	}
}

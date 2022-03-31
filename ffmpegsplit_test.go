package ffmpegsplit

import (
	"encoding/json"
	"testing"
)

func TestDecode(t *testing.T) {
	var probeOut FFProbeOutput
	err := json.Unmarshal([]byte(chapters_json), &probeOut)
	if err != nil {
		t.Fatalf("Failed to decode chapters JSON: %v", err)
	}
	if len(probeOut.Chapters) != 3 {
		t.Fatalf("Expected 3 chapters, got %v", len(probeOut.Chapters))
	}
}

// test data
var chapters_json string = `
{
    "chapters": [
        {
            "id": 0,
            "time_base": "1/1000",
            "start": 0,
            "start_time": "0.000000",
            "end": 20000,
            "end_time": "20.000000",
            "tags": {
                "title": "It All Started With a Simple BEEP"
            }
        },
        {
            "id": 1,
            "time_base": "1/1000",
            "start": 20000,
            "start_time": "20.000000",
            "end": 40000,
            "end_time": "40.000000",
            "tags": {
                "title": "All You Can BEEP Buffee"
            }
        },
        {
            "id": 2,
            "time_base": "1/1000",
            "start": 40000,
            "start_time": "40.000000",
            "end": 60000,
            "end_time": "60.000000",
            "tags": {
                "title": "The Final Beep"
            }
        }
    ]
}
`

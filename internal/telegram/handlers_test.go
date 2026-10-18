package telegram

import "testing"

func TestStartPayload(t *testing.T) {
	tests := map[string]string{
		"/start":                   "",
		"/start Ab3xK9pQr2Zt":      "Ab3xK9pQr2Zt",
		"  /start   Ab3xK9pQr2Zt ": "Ab3xK9pQr2Zt",
		"/start@radar_bot code":    "code",
	}
	for text, want := range tests {
		if got := startPayload(text); got != want {
			t.Errorf("startPayload(%q) = %q, want %q", text, got, want)
		}
	}
}

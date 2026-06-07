package middleware

import "testing"

func TestOriginAllowed(t *testing.T) {
	allowed := []string{
		"https://www.connectoo.online",
		"https://admin.connectoo.online",
	}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "exact match", origin: "https://www.connectoo.online", want: true},
		{name: "trailing slash on request", origin: "https://www.connectoo.online/", want: true},
		{name: "trailing slash in allow list", origin: "https://admin.connectoo.online", want: true},
		{name: "wrong host", origin: "https://evil.example", want: false},
		{name: "empty origin", origin: "", want: false},
		{name: "http vs https", origin: "http://www.connectoo.online", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := originAllowed(allowed, tt.origin); got != tt.want {
				t.Fatalf("originAllowed(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

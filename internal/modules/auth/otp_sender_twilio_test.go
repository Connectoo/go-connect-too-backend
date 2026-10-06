package auth

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// roundTripFunc lets a plain function act as an http.RoundTripper so tests can
// capture the outgoing request and return a canned response without any network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestToE164(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "already e164", in: "+919876543210", want: "+919876543210"},
		{name: "no plus", in: "919876543210", want: "+919876543210"},
		{name: "with spaces and plus", in: "+91 98765 43210", want: "+919876543210"},
		{name: "double zero prefix", in: "0091 98765", want: "+9198765"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toE164(tt.in); got != tt.want {
				t.Errorf("toE164(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTwilioOTPSenderSend(t *testing.T) {
	const (
		accountSID = "AC0123456789abcdef"
		authToken  = "super-secret-token"
		fromNumber = "+15551234567"
		code       = "654321"
	)

	t.Run("success posts expected request", func(t *testing.T) {
		var captured *http.Request
		var capturedBody string
		client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			captured = req
			b, _ := io.ReadAll(req.Body)
			capturedBody = string(b)
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(`{"sid":"SM1","status":"queued"}`)),
				Header:     make(http.Header),
			}, nil
		})}

		sender := NewTwilioOTPSender(accountSID, authToken, fromNumber, "", "Code: {code}", client, nil)
		if err := sender.Send("+91 98765 43210", code); err != nil {
			t.Fatalf("Send() error = %v, want nil", err)
		}

		if captured == nil {
			t.Fatal("expected a request to be captured")
		}
		if !strings.Contains(captured.URL.Path, accountSID) {
			t.Errorf("URL path %q should contain account SID", captured.URL.Path)
		}
		if !strings.HasSuffix(captured.URL.Path, "/Messages.json") {
			t.Errorf("URL path %q should end with /Messages.json", captured.URL.Path)
		}
		user, pass, ok := captured.BasicAuth()
		if !ok {
			t.Error("expected Basic auth header to be present")
		}
		if user != accountSID || pass != authToken {
			t.Error("Basic auth credentials did not match account SID / auth token")
		}

		form, err := url.ParseQuery(capturedBody)
		if err != nil {
			t.Fatalf("parse form body: %v", err)
		}
		if got := form.Get("To"); got != "+919876543210" {
			t.Errorf("To = %q, want +919876543210 (E.164 with leading +)", got)
		}
		if got := form.Get("From"); got != fromNumber {
			t.Errorf("From = %q, want %q", got, fromNumber)
		}
		if got := form.Get("Body"); !strings.Contains(got, code) {
			t.Errorf("Body = %q, want it to contain the code", got)
		}
	})

	t.Run("messaging service sid preferred over from", func(t *testing.T) {
		var capturedBody string
		client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			capturedBody = string(b)
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`)), Header: make(http.Header)}, nil
		})}

		sender := NewTwilioOTPSender(accountSID, authToken, fromNumber, "MG999", "", client, nil)
		if err := sender.Send("919876543210", code); err != nil {
			t.Fatalf("Send() error = %v", err)
		}
		form, _ := url.ParseQuery(capturedBody)
		if got := form.Get("MessagingServiceSid"); got != "MG999" {
			t.Errorf("MessagingServiceSid = %q, want MG999", got)
		}
		if form.Get("From") != "" {
			t.Errorf("From should be empty when MessagingServiceSid is set, got %q", form.Get("From"))
		}
	})

	t.Run("twilio 4xx returns wrapped error without leaking secrets", func(t *testing.T) {
		client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"code":21211,"message":"Invalid 'To' Phone Number"}`)),
				Header:     make(http.Header),
			}, nil
		})}

		sender := NewTwilioOTPSender(accountSID, authToken, fromNumber, "", "Code: {code}", client, nil)
		err := sender.Send("+919876543210", code)
		if err == nil {
			t.Fatal("Send() expected an error on 4xx, got nil")
		}
		msg := err.Error()
		if strings.Contains(msg, code) {
			t.Error("error must not leak the OTP code")
		}
		if strings.Contains(msg, authToken) {
			t.Error("error must not leak the auth token")
		}
		if !strings.Contains(msg, "400") {
			t.Errorf("error = %q, want it to mention the HTTP status", msg)
		}
	})
}

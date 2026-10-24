package server

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestMonitorResponsesAreBoundedAndErrorsOmitCredentials(t *testing.T) {
	if _, err := readMonitorResponse(strings.NewReader(strings.Repeat("x", (4<<20)+1))); err == nil {
		t.Fatal("oversized response accepted")
	}
	if data, err := readMonitorResponse(strings.NewReader("synthetic")); err != nil || string(data) != "synthetic" {
		t.Fatal("normal response rejected")
	}
	message := transportError(&url.Error{Op: "Post", URL: "https://monitor.example/?token=synthetic-secret", Err: errors.New("synthetic failure")}).Error()
	if strings.Contains(message, "synthetic-secret") {
		t.Fatal("token in error log")
	}
	if strings.Contains((&httpStatusError{StatusCode: 400, Body: "synthetic-secret"}).Error(), "synthetic-secret") {
		t.Fatal("response body in error log")
	}
}

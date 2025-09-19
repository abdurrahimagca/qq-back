package server

import (
	"time"
)

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
func timeStrPtr(t time.Time) *string {
	return stringPtr(t.Format(time.RFC3339))
}
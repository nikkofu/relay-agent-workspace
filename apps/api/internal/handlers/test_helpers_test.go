package handlers

import (
	"regexp"
	"testing"
)

func assertPrefixedUUID(t *testing.T, value, prefix string) {
	t.Helper()

	pattern := "^" + regexp.QuoteMeta(prefix) + `-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		t.Fatalf("failed to match %s id %q: %v", prefix, value, err)
	}
	if !matched {
		t.Fatalf("expected %s id to use uuid format, got %q", prefix, value)
	}
}

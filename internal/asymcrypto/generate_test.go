package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	err := GeneratePair()
	if err != nil {
		t.Errorf(`GeneratePair("") = %v, want "", error`, err)
	}
}

package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	t.Helper()
	t.Parallel()

	err := GeneratePair()
	if err != nil {
		t.Errorf(`TestGeneratePair("") = %v, want "", error`, err)
	}
}

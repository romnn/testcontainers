package main

import (
	"testing"
)

func TestCli(t *testing.T) {
	out := run()
	expected := "This is an example!"
	if out != expected {
		t.Errorf("Got %s but expected %s", out, expected)
	}
}

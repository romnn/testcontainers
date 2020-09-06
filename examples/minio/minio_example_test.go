package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMongoExample(t *testing.T) {
	t.Parallel()
	log.SetLevel(log.ErrorLevel)
	out := run()
	expected := "hello world"
	if out != expected {
		t.Errorf("Got %q but expected %q", out, expected)
	}
}

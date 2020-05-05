package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestRedisExample(t *testing.T) {
	t.Parallel()
	log.SetLevel(log.ErrorLevel)
	out := run()
	expected := "Hello World!"
	if out != expected {
		t.Errorf("Got %s but expected %s", out, expected)
	}
}

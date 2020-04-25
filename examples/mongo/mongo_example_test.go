package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMongoExample(t *testing.T) {
	t.Parallel()
	log.SetLevel(log.ErrorLevel)
	out := run()
	expected := 32 // 50 - 18 = 32
	if out != expected {
		t.Errorf("Got %d but expected %d", out, expected)
	}
}

package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestKafkaExample(t *testing.T) {
	t.Parallel()
	log.SetLevel(log.ErrorLevel)
	out := run()
	expected := "Consuming and producing works!"
	if out != expected {
		t.Errorf("Got %s but expected %s", out, expected)
	}
}

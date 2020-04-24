package testcontainers

import "testing"

func TestShout(t *testing.T) {
	if Shout("Test") != "Test!" {
		t.Errorf("Got %s but want \"Test!\"", Shout("Test"))
	}
	if Shout("") != "!" {
		t.Errorf("Got %s but want \"!\"", Shout(""))
	}
}

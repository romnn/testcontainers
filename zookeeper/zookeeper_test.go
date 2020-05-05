package zookeeper

import (
	"context"
	"testing"
)

// TestZookeeperContainer ...
func TestZookeeperContainer(t *testing.T) {
	t.Parallel()
	// Only test zookeeper container starts without errors, functionality is tested with kafka
	zkC, _, err := StartZookeeperContainer(ContainerOptions{})
	if err != nil {
		t.Fatalf("Failed to start Zookeeper container: %v", err)
	}
	defer zkC.Terminate(context.Background())
}

package testcontainers

import "sync"

// Version is incremented using bump2version
const Version = "0.0.1"

var clientMux sync.Mutex

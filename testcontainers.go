package testcontainers

import "sync"

// Version is incremented using bump2version
const Version = "0.1.3"

var clientMux sync.Mutex

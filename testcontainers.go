package testcontainers

import "sync"

// Version is incremented using bump2version
const Version = "0.1.9"

// ClientMux ...
var ClientMux sync.Mutex

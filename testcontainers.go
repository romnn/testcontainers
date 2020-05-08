package testcontainers

import "sync"

// Version is incremented using bump2version
const Version = "0.1.8"

// ClientMux ...
var ClientMux sync.Mutex

your-api-client/ (or Go-Enterprise-APIs/)
├── go.mod
├── go.sum
├── client.go               // Main client struct and constructor (e.g., CircularClient)
├── config.go               // Configuration options for the client
├── errors.go               // Custom error types
├── models/                 // Data structures for API requests and responses
│   ├── ccertificate.go
│   └── account.go
├── services/               // API resource/service specific logic
│   └── cepaccount.go
├── internal/               // Internal utility functions, not exported
│   ├── http.go             // Helper for making HTTP requests
│   ├── utils.go            // Helper utilities (from lib/utils)
│   └── crypto.go           // Cryptographic utilities (from lib/cepaccount/crypto.go)
├── examples/               // Usage examples
│   └── main.go             // Main example
└── ...

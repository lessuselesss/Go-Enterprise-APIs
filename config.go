package circular

// Config holds the configuration options for the Circular client.
type Config struct {
	NagURL      string
	NetworkNode string
	Blockchain  string
	IntervalSec int
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		NagURL:      "https://nag.circular.io/", // Assuming this default from internal/utils.go
		Blockchain:  "Circular",                 // Assuming this default from internal/utils.go
		IntervalSec: 2,                          // Assuming this default from services/account.go
		// NetworkNode is intentionally left empty as it might be set later
	}
}

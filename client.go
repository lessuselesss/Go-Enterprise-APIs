package circular

import (
	"circular-api/services"
)

// CircularClient is the main client for interacting with the Circular API.
type CircularClient struct {
	Config  *Config
	Account *services.CEPAccount
}

// NewCircularClient creates a new instance of the CircularClient.
func NewCircularClient(config *Config) *CircularClient {
	if config == nil {
		config = NewConfig()
	}
	return &CircularClient{
		Config:  config,
		Account: services.NewCEPAccount(), // Initialize the Account service
	}
}

// Note: Methods that delegate to services will be added later.

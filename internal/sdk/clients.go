package sdk

import (
	"context"
	"fmt"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

type Clients struct {
	Region string

	GlobalClient   *secapi.GlobalClient
	RegionalClient *secapi.RegionalClient
}

type Config struct {
	Token  string
	Region string

	GlobalProviders *ConfigGlobalProviders
}

type ConfigGlobalProviders struct {
	RegionV1        string
	AuthorizationV1 string
}

func InitClients(ctx context.Context, config *Config) (*Clients, error) {
	regionV1Endpoint := config.GlobalProviders.RegionV1
	authV1Endpoint := config.GlobalProviders.AuthorizationV1

	// Initialize global client
	globalClient, err := secapi.NewGlobalClient(&secapi.GlobalConfig{
		AuthToken: config.Token,
		Endpoints: secapi.GlobalEndpoints{
			RegionV1:        regionV1Endpoint,
			AuthorizationV1: authV1Endpoint,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create global client: %w", err)
	}

	// Initialize regional client
	regionalClient, err := globalClient.NewRegionalClient(ctx, config.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create regional client: %w", err)
	}

	return &Clients{
		GlobalClient:   globalClient,
		RegionalClient: regionalClient,
		Region:         config.Region,
	}, nil
}

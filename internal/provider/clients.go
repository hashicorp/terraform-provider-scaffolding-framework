package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"
)

const (
	defaultRetryDelay       = 30 * time.Second
	defaultRetryInterval    = 10 * time.Second
	defaultRetryMaxAttempts = 5
)

type clients struct {
	Tenant string
	Region string

	RetryDelay       time.Duration
	RetryInterval    time.Duration
	RetryMaxAttempts int

	GlobalClient   *secapi.GlobalClient
	RegionalClient *secapi.RegionalClient
}

type clientConfig struct {
	Token  string
	Tenant string
	Region string

	RetryDelay       time.Duration
	RetryInterval    time.Duration
	RetryMaxAttempts int

	GlobalProviders *clientConfigGlobalProviders
}

type clientConfigGlobalProviders struct {
	RegionV1        string
	AuthorizationV1 string
}

func initClients(ctx context.Context, config *clientConfig) (*clients, error) {
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

	return &clients{
		Tenant: config.Tenant,
		Region: config.Region,

		RetryDelay:       config.RetryDelay,
		RetryInterval:    config.RetryInterval,
		RetryMaxAttempts: config.RetryMaxAttempts,

		GlobalClient:   globalClient,
		RegionalClient: regionalClient,
	}, nil
}

package acctest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/provider"

	"github.com/eu-sovereign-cloud/go-sdk/secapi"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	testAccToken        = "test"
	testAccTenant       = "seca"
	testAccRegion       = "region"
	testAccEndpointReg  = "http://172.18.0.2:30081/providers/seca.region"
	testAccEndpointAuth = "http://172.18.0.2:30081/providers/seca.authorization"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seca": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "seca" {
  token  = %q
  tenant = %q
  region = %q
  global_providers = {
    region_v1        = %q,
    authorization_v1 = %q
  }
}`, testAccToken, testAccTenant, testAccRegion, testAccEndpointReg, testAccEndpointAuth)
}

func testAccRegionalClient(ctx context.Context) (*secapi.RegionalClient, error) {
	globalClient, err := secapi.NewGlobalClient(&secapi.GlobalConfig{
		AuthToken: testAccToken,
		Endpoints: secapi.GlobalEndpoints{
			RegionV1:        testAccEndpointReg,
			AuthorizationV1: testAccEndpointAuth,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create global client: %w", err)
	}

	regionalClient, err := globalClient.NewRegionalClient(ctx, testAccRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to create regional client: %w", err)
	}

	return regionalClient, nil
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("{\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "    %s = %q\n", k, labels[k])
	}
	b.WriteString("  }")
	return b.String()
}

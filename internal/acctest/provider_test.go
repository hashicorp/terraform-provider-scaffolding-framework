package acctest

import (
	"testing"

	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seca": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccProviderConfig() string {
	return `
provider "seca" {
  token  = "test"
  tenant = "seca"
  region = "region"
  global_providers = {
    region_v1        = "http://172.18.0.2:30081/providers/seca.region",
    authorization_v1 = "http://172.18.0.2:30081/providers/seca.authorization"
  }
}`
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
}

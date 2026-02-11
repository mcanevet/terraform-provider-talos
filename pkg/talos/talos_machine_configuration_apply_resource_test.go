// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package talos_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccTalosMachineConfigurationApplyResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				Source:            "dmacvicar/libvirt",
				VersionConstraint: "= 0.8.3",
			},
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTalosMachineConfigurationApplyResourceConfig("talos", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "apply_mode", "auto"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "node"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "endpoint"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.ca_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_key"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration_input"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.#", "1"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.0", "\"machine\":\n  \"install\":\n    \"disk\": \"/dev/vda\"\n"),
				),
			},
			// ensure there is no diff
			{
				Config:   testAccTalosMachineConfigurationApplyResourceConfig("talos", rName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccTalosMachineConfigurationApplyResourceAutoStaged tests the "staged_if_needing_reboot" apply mode.
//
// Note on local vs CI environment:
// During local development, the node IP was sometimes unknown during the plan phase,
// preventing the dry-run from being performed. However, in CI, the libvirt setup
// allows the node IP to be known immediately, enabling the dry-run to execute.
// Since the configuration requires a reboot, the dry-run correctly resolves to
// "staged" mode to prevent uncontrolled reboots.
func TestAccTalosMachineConfigurationApplyResourceAutoStaged(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"libvirt": {
				Source:            "dmacvicar/libvirt",
				VersionConstraint: "= 0.8.3",
			},
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTalosMachineConfigurationApplyResourceConfigWithAutoStaged("talos", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.staged_if_needing_reboot", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.staged_if_needing_reboot", "apply_mode", "staged_if_needing_reboot"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.staged_if_needing_reboot", "resolved_apply_mode", "staged"),
				),
			},
		},
	})
}

func TestAccTalosMachineConfigurationApplyResourceUpgrade(t *testing.T) {
	// ref: https://github.com/hashicorp/terraform-plugin-testing/pull/118
	t.Skip("skipping until TF test framework has a way to remove state resource")

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			// create TF config with v0.1.2 of the talos provider
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"talos": {
						VersionConstraint: "=0.1.2",
						Source:            "siderolabs/talos",
					},
					"libvirt": {
						Source:            "dmacvicar/libvirt",
						VersionConstraint: "= 0.8.3",
					},
				},
				Config: testAccTalosMachineConfigurationApplyResourceConfigV0("talosv1", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("talos_client_configuration", "this"),
					resource.TestCheckNoResourceAttr("talos_machine_configuration_controlplane", "this"),
				),
			},
			// now test state migration with the latest version of the provider
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"libvirt": {
						Source:            "dmacvicar/libvirt",
						VersionConstraint: "= 0.8.3",
					},
				},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccTalosMachineConfigurationApplyResourceConfigV1("talos", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "id", "machine_configuration_apply"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "apply_mode", "auto"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "node"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "endpoint"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.ca_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_certificate"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "client_configuration.client_key"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration_input"),
					resource.TestCheckResourceAttrSet("talos_machine_configuration_apply.this", "machine_configuration"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.#", "1"),
					resource.TestCheckResourceAttr("talos_machine_configuration_apply.this", "config_patches.0", "\"machine\":\n  \"install\":\n    \"disk\": \"/dev/vda\"\n"),
				),
			},
			// ensure there is no diff
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"libvirt": {
						Source:            "dmacvicar/libvirt",
						VersionConstraint: "= 0.8.3",
					},
				},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   testAccTalosMachineConfigurationApplyResourceConfigV1("talos", rName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccTalosMachineConfigurationApplyResourceConfig(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	return config.render()
}

func testAccTalosMachineConfigurationApplyResourceConfigV0(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	return config.render()
}

func testAccTalosMachineConfigurationApplyResourceConfigV1(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	return config.render()
}

func TestAccTalosMachineConfigurationApplyValidateNeitherInput(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "talos_machine_configuration_apply" "this" {
  node = "10.5.0.2"
  client_configuration = {
    ca_certificate     = "ca"
    client_certificate = "cert"
    client_key         = "key"
  }
}
`,
				ExpectError: regexp.MustCompile(`Missing machine configuration input`),
			},
		},
	})
}

func TestAccTalosMachineConfigurationApplyValidateBothInputs(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "talos_machine_configuration_apply" "this" {
  node = "10.5.0.2"
  client_configuration = {
    ca_certificate     = "ca"
    client_certificate = "cert"
    client_key         = "key"
  }
  machine_configuration_input    = "some-config"
  machine_configuration_input_wo = "some-config"
}
`,
				ExpectError: regexp.MustCompile(`Conflicting machine configuration input`),
			},
		},
	})
}

func TestAccTalosMachineConfigurationApplyValidateWriteOnlyInput(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "talos_machine_secrets" "this" {}

data "talos_machine_configuration" "this" {
  cluster_name     = "test-cluster"
  cluster_endpoint = "https://10.0.0.1:6443"
  machine_type     = "controlplane"
  machine_secrets  = talos_machine_secrets.this.machine_secrets
}

resource "talos_machine_configuration_apply" "this" {
  node = "10.5.0.2"
  client_configuration = {
    ca_certificate     = "ca"
    client_certificate = "cert"
    client_key         = "key"
  }
  machine_configuration_input_wo = data.talos_machine_configuration.this.machine_configuration
}
`,
				// Validation passes (no "Missing" or "Conflicting" error).
				// The error comes from the apply phase, proving that the
				// write-only input was accepted and plan succeeded.
				ExpectError: regexp.MustCompile(`Error (applying|converting) config`),
			},
		},
	})
}

func testAccTalosMachineConfigurationApplyResourceConfigWithAutoStaged(providerName, rName string) string {
	config := dynamicConfig{
		Provider:        providerName,
		ResourceName:    rName,
		WithApplyConfig: true,
		WithBootstrap:   false,
	}

	baseConfig := config.render()

	return baseConfig + `
resource "talos_machine_configuration_apply" "staged_if_needing_reboot" {
  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.this.machine_configuration
  node                        = libvirt_domain.cp.network_interface[0].addresses[0]
  apply_mode                  = "staged_if_needing_reboot"
  config_patches = [
    yamlencode({
      machine = {
        files = [
          {
            path        = "/var/etc/example-config.yaml"
            permissions = 420  # 0644 in octal
            op          = "create"
            content     = "example: staged_if_needing_reboot test"
          }
        ]
      }
    }),
  ]
}
`
}

func TestAccTalosMachineConfigurationApplyWithUnknownClientConfig(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "talos_machine_secrets" "this" {}

data "talos_machine_configuration" "this" {
  cluster_name     = "test-cluster"
  cluster_endpoint = "https://10.5.0.2:6443"
  machine_type     = "controlplane"
  machine_secrets  = talos_machine_secrets.this.machine_secrets
}

resource "talos_machine_configuration_apply" "this" {
  # client_configuration comes from a computed attribute, so it's unknown during plan
  client_configuration        = talos_machine_secrets.this.client_configuration
  machine_configuration_input = data.talos_machine_configuration.this.machine_configuration
  node                        = "10.5.0.2"

  # This apply_mode triggers reboot prevention logic which needs to handle unknown client_config
  apply_mode = "staged_if_needing_reboot"
}
`,
				// PlanOnly ensures we test the plan phase where client_configuration is unknown
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}


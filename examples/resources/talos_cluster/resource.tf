locals {
  talos_version      = "v1.13.2"
  kubernetes_version = "v1.36.0"
}

resource "talos_machine_secrets" "this" {}

ephemeral "talos_machine_configuration" "this" {
  cluster_name       = "example-cluster"
  machine_type       = "controlplane"
  cluster_endpoint   = "https://10.5.0.2:6443"
  machine_secrets    = talos_machine_secrets.this.machine_secrets
  talos_version      = local.talos_version
  kubernetes_version = local.kubernetes_version
  config_patches = [
    yamlencode({
      machine = {
        install = {
          disk  = "/dev/sda"
          image = "ghcr.io/siderolabs/installer:${local.talos_version}"
        }
      }
    })
  ]
}

resource "talos_machine" "this" {
  node                     = "10.5.0.2"
  client_configuration     = talos_machine_secrets.this.client_configuration
  machine_configuration_wo = ephemeral.talos_machine_configuration.this.machine_configuration
  image                    = "ghcr.io/siderolabs/installer:${local.talos_version}"
}

resource "talos_cluster" "this" {
  node                 = talos_machine.this.node
  client_configuration = talos_machine_secrets.this.client_configuration
  kubernetes_version   = local.kubernetes_version
}

data "talos_cluster_kubeconfig" "this" {
  client_configuration = talos_machine_secrets.this.client_configuration
  node                 = "10.5.0.2"
  depends_on           = [talos_cluster.this]
}

terraform {
  required_providers {
    azurerm = {
      source  = "terraform-registry.dev.local/hashicorp/azurerm"
      version = "=2.97.0"
    }
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
}

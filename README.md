# terraform-registry-reverse-proxy

This is used to proxy internal requests back to Terraform's Registry.

## What is this?

Hashicorp for some reason either hasn't prioritized or outright refuses to make
it easy to ingest Terraform providers and modules like you might traditionally
see in an Artifactory so that you don't have to cache the plugins and unzip them.
Instead just let the `terraform` cli natively fetch things on-demand in an
environment where direct internet access is not possible.

This application is intended to be put behind a web server, e.g. [NGINX][1], [Caddy][2],
[Apache][3], etc.

## How does it work?

The tiny reverse proxy app is really quite simple and does two things:

1. Proxies requests to https://registry.terraform.io and https://releases.hashicorp.com (optionally, if not using external storage)
2. Re-write response bodies to update where Artifacts should be fetched from (configurable).
   - For example, it will replace:
     - original url: `https://releases.hashicorp.com/terraform-provider-azurerm/2.97.0/terraform-provider-azurerm_2.97.0_darwin_amd64.zip`
     - re-written to: `https://hashicorp-releases.company.com/terraform-provider-azurerm/2.97.0/terraform-provider-azurerm_2.97.0_darwin_amd64.zip`
     - or with Artifactory re-written to: `https://artifactory.company.com/artifactory/hashicorp-releases/terraform-provider-azurerm/2.97.0/terraform-provider-azurerm_2.97.0_darwin_amd64.zip`

## Requirements

- web server binding on TCP/443
- ssl certificate(s) that is/are trusted by the client where `terraform` is
  being run
- dns record dedicated to terraform registry proxy (e.g. `terraform-registry.company.com`)
- update sources in your terraform configurations

### Optionally

- artifact storage (e.g. Artifactory)
- dns record dedicated to hashicorp releases proxy (e.g. `hashicorp-releases.company.com`)

## Usage

Two possible usages:

1. Without external artifact storage
2. With external artifact storage

Read each section below for more details. We will use [Artifactory][4] as our example
artifact storage and the Caddy web server for our examples as well, but any web
server or artifact storage should work.

### Without external artifact storage

![with artifact storage](/docs/diagrams/without-artifact-storage.drawio.png?raw=true)

```bash
./terraform-registry-reverse-proxy -registry-proxy-host terraform-registry.company.com \
                                   -release-proxy-host hashicorp-releases.company.com
```

### With external artifact storage

![with artifact storage](/docs/diagrams/with-artifact-storage.drawio.png?raw=true)

```bash
./terraform-registry-reverse-proxy -registry-proxy-host terraform-registry.company.com \
                                   -release-proxy-host artifactory.company.com \
                                   -release-proxy-path-prefix /artifactory/hashicorp-releases
```

### Update sources in Terraform configurations

After you have your infrastructure setup you need to update your Terraform
configuration so it knows to pull dependencies through the proxy.

Say for example this is your original configuration:

```terraform
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=2.97.0"
    }
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
}
```

You would update it to this

```terraform
terraform {
  required_providers {
    azurerm = {
      source  = "terraform-registry.company.com/hashicorp/azurerm"
      version = "=2.97.0"
    }
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
}
```

[1]: https://nginx.org/en/
[2]: https://caddyserver.com/
[3]: https://httpd.apache.org/
[4]: https://jfrog.com/artifactory/

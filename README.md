# terraform-registry-proxy

This app is useful if you run Terraform in an offline / airgappped / no internet connected environment.

## What is this?

Hashicorp for some reason either hasn't prioritized or outright refuses to make
it easy to ingest Terraform providers and modules like you might traditionally
see in an Artifactory so that you don't have to cache the plugins and unzip them.
Instead just let the `terraform` cli natively fetch things on-demand in an
environment where direct internet access is not possible.

This application is intended to be put behind a web server, e.g. [NGINX][1], [Caddy][2],
[Apache][3], etc.

## How does it work?

The tiny proxy app is really quite simple and does two things:

1. Proxies requests to https://registry.terraform.io and https://releases.hashicorp.com (optionally, if not using external artifact storage)
2. Re-write response bodies to update where Artifacts should be fetched from (configurable).
   - For example, it will replace:
     - original url: `https://releases.hashicorp.com/terraform-provider-azurerm/2.97.0/terraform-provider-azurerm_2.97.0_darwin_amd64.zip`
     - re-written to: `https://hashicorp-releases.company.com/terraform-provider-azurerm/2.97.0/terraform-provider-azurerm_2.97.0_darwin_amd64.zip`
     - or with Artifactory re-written to: `https://artifactory.company.com/artifactory/hashicorp-releases/terraform-provider-azurerm/2.97.0/terraform-provider-azurerm_2.97.0_darwin_amd64.zip`

## Requirements

- web server
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

These diagrams are not intended to be recommendations for specific architectures
but simply showing you examples of possible ways to set it up to make it easier
to get familiar with how it works.

### Without external artifact storage

In this scenario both https://registry.terraform.io and
https://releases.hashicorp.com are proxied through this app.

You will need to setup two DNS records pointing to the web server where
`terraform-registry-proxy` is running, i.e.

- `terraform-registry.company.com`
- `hashicorp-releases.company.com`

![with artifact storage](/docs/diagrams/without-artifact-storage.drawio.png?raw=true)

```bash
./terraform-registry-proxy -registry-proxy-host terraform-registry.company.com \
                                   -release-proxy-host hashicorp-releases.company.com
```

### With external artifact storage

In this scenario only https://registry.terraform.io is proxied through this app.

You will need to setup one DNS record pointing to the web server where
`terraform-registry-proxy` is running, i.e.

- `terraform-registry.company.com`

It also assumes you're already proxying / mirroring https://releases.hashicorp.com.

![with artifact storage](/docs/diagrams/with-artifact-storage.drawio.png?raw=true)

```bash
./terraform-registry-proxy -registry-proxy-host terraform-registry.company.com \
                                   -release-proxy-host artifactory.company.com \
                                   -release-proxy-path-prefix /artifactory/hashicorp-releases
```

This assumes you've configured a generic remote repository named
`hashicorp-releases` for https://registry.terraform.io in your Artifactory
instance.

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

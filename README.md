# terraform-registry-reverse-proxy

This is used to proxy internal requests back to Terraform's Registry.

## What is this?

Hashicorp for some reason either hasn't prioritized or outright refuses to make
it easy to ingest Terraform providers and modules like you might traditionally
see in an Artifactory so that you don't have to cache the plugins and unzip them
but just let the `terraform` cli natively fetch things on-demand in an
environment where direct internet access is not possible.

This application is intended to be put behind a web server, e.g. NGINX, Caddy,
Apache, etc. This gives you flexibility to use whichever webserver you want.

## Requirements

- Any web server
- SSL certificates

## Usage

Two possible usages:

### Without Artifactory

### With Artifactory

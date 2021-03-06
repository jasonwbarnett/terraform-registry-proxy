package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
)

// WebReverseProxyConfiguration is a coniguration for the ReverseProxy
type WebReverseProxyConfiguration struct {
	RegistryProxyHost string
	ReleaseProxyHost  string
	ReleasePathPrefix string
}

var (
	registryHost      string
	releaseHost       string
	releasePathPrefix string
	httpAddress       string
)

func init() {
	flag.StringVar(&registryHost, "registry-proxy-host", "", "FQDN of registry proxy host [Required]")
	flag.StringVar(&releaseHost, "release-proxy-host", "", "FQDN of release proxy host [Required]")
	flag.StringVar(&releasePathPrefix, "release-proxy-path-prefix", "", "The prefix path to prepend to any release artifact paths. This might be /artifactory/hashicorp-releases")
	flag.StringVar(&httpAddress, "http-address", ":8555", "HTTP address to listen on, e.g. :8080 or 127.0.0.1:8080")
	flag.Parse()

	if registryHost == "" {
		fmt.Printf("You must provide a -registry-proxy-host value\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if releaseHost == "" {
		fmt.Printf("You must provide a -release-proxy-host value\n\n")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	config := &WebReverseProxyConfiguration{
		RegistryProxyHost: registryHost,
		ReleaseProxyHost:  releaseHost,
		ReleasePathPrefix: releasePathPrefix,
	}
	proxy := config.NewWebReverseProxy()
	http.Handle("/", handlers.LoggingHandler(os.Stdout, proxy))

	// Start the server
	http.ListenAndServe(httpAddress, nil)
}

// This replaces all occurrences of http://releases.hashicorp.com with
// config.ReleaseProxyHost in the response body
func (config *WebReverseProxyConfiguration) rewriteBody(resp *http.Response) (err error) {
	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
		resp.Uncompressed = true
		defer reader.Close()
	default:
		reader = resp.Body
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if err = resp.Body.Close(); err != nil {
		return err
	}

	replacement := fmt.Sprintf("https://%s%s", config.ReleaseProxyHost, config.ReleasePathPrefix)

	b = bytes.ReplaceAll(b, []byte("https://releases.hashicorp.com"), []byte(replacement)) // releases
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

func (config *WebReverseProxyConfiguration) NewWebReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		if req.Host == config.RegistryProxyHost {
			req.URL.Scheme = "https"
			req.URL.Host = "registry.terraform.io"
			req.Host = "registry.terraform.io"
			req.Header.Set("User-Agent", "Terraform/1.1.7")
			req.Header.Set("X-Terraform-Version", "1.1.7")
		} else if req.Host == config.ReleaseProxyHost {
			req.URL.Scheme = "https"
			req.URL.Host = "releases.hashicorp.com"
			req.Host = "releases.hashicorp.com"
			req.Header.Set("User-Agent", "Terraform/1.1.7")
		}
	}

	responseDirector := func(res *http.Response) error {
		if server := res.Header.Get("Server"); strings.HasPrefix(server, "terraform-registry") {
			if err := config.rewriteBody(res); err != nil {
				fmt.Println("Error rewriting body!")
				return err
			}
		}

		if location := res.Header.Get("Location"); location != "" {
			url, err := url.ParseRequestURI(location)
			if err != nil {
				fmt.Println("Error!")
				return err
			}

			// Override redirect url Host with ProxyHost
			url.Host = config.RegistryProxyHost

			res.Header.Set("Location", url.String())
			res.Header.Set("X-Reverse-Proxy", "terraform-registry-proxy")
		}
		return nil
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: responseDirector,
	}
}

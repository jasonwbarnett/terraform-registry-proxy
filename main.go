package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// WebReverseProxyConfiguration is a coniguration for the ReverseProxy
type WebReverseProxyConfiguration struct {
	ProxyHost string
}

func main() {
	config := &WebReverseProxyConfiguration{
		ProxyHost: "127.0.0.1:8080",
	}
	proxy := NewWebReverseProxy(config)
	http.Handle("/", proxy)

	// Start the server
	http.ListenAndServe(":8080", nil)
}

func pullDomainAndPath(a string) (domain string, path string) {
	data := strings.Split(a, "/")
	domain = data[1]
	path = "/" + strings.Join(data[2:], "/")

	return domain, path
}

func convertURLToProxy(config *WebReverseProxyConfiguration, u *url.URL) string {
	newURL := "http://" + config.ProxyHost + u.Path
	if u.RawQuery != "" {
		newURL = newURL + "?" + u.RawQuery
	}

	return newURL
}

func NewWebReverseProxy(config *WebReverseProxyConfiguration) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = "registry.terraform.io"
		req.Host = req.URL.Host
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	responseDirector := func(res *http.Response) error {
		if location := res.Header.Get("Location"); location != "" {
			url, err := url.ParseRequestURI(location)
			if err != nil {
				fmt.Println("Error!")
				return err
			}

			// Override redirect url Host with ProxyHost
			url.Host = config.ProxyHost

			res.Header.Set("Location", url.String())
			res.Header.Set("X-Reverse-Proxy", "terraform-registry-reverse-proxy")
		}
		return nil
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: responseDirector,
	}
}

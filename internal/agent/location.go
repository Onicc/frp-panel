package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var publicIPEndpoints = []string{
	"https://api.ipify.org",
	"https://ipv4.icanhazip.com",
	"https://ifconfig.me/ip",
}

func publicIP(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
		return ""
	}
	return ip.String()
}

// ProbePublicIP bypasses HTTP_PROXY/HTTPS_PROXY. A transparent network or
// TUN proxy may still intercept packets; the console supports manual override.
func ProbePublicIP(ctx context.Context) (string, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	client := &http.Client{Timeout: 4 * time.Second, Transport: transport}
	defer transport.CloseIdleConnections()
	for _, endpoint := range publicIPEndpoints {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 128))
		resp.Body.Close()
		if readErr == nil && resp.StatusCode == http.StatusOK {
			if ip := publicIP(string(body)); ip != "" {
				return ip, nil
			}
		}
	}
	return "", fmt.Errorf("all direct public IP probes failed")
}

func ReportPublicIP(ctx context.Context, apiURL, clientID, secret, ip string, insecureSkipVerify bool) error {
	if publicIP(ip) == "" {
		return fmt.Errorf("invalid public IP")
	}
	base, err := url.Parse(apiURL)
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" {
		return fmt.Errorf("invalid Master API URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/api/v2/agent/location"
	base.RawQuery = ""
	base.Fragment = ""
	body, err := json.Marshal(map[string]string{"ip": ip})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FRP-Panel-Client-ID", clientID)
	req.Header.Set("X-FRP-Panel-Client-Secret", secret)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecureSkipVerify} // #nosec G402 -- explicit Agent configuration
	client := &http.Client{Timeout: 10 * time.Second, Transport: transport}
	defer transport.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Master rejected location report: %s", resp.Status)
	}
	return nil
}

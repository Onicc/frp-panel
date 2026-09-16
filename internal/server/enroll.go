package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type EnrollmentResult struct {
	ServerID string `json:"serverId"`
	Secret   string `json:"secret"`
}

type ResolveOptions struct {
	ConfigPath      string
	EnrollmentToken string
	APIURL          string
	RPCURL          string
	FallbackID      string
	FallbackSecret  string
	Insecure        bool
}

func Resolve(options ResolveOptions) (Config, error) {
	if options.ConfigPath == "" {
		return Config{}, fmt.Errorf("server config path is required")
	}
	if _, err := os.Stat(options.ConfigPath); err == nil {
		return ReadConfig(options.ConfigPath)
	} else if !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("inspect server config: %w", err)
	}

	result := Config{
		Version:     ConfigVersion,
		Master:      Master{APIURL: strings.TrimSpace(options.APIURL), RPCURL: strings.TrimSpace(options.RPCURL)},
		Credentials: Credentials{ServerID: strings.TrimSpace(options.FallbackID), Secret: options.FallbackSecret},
		TLS:         TLS{InsecureSkipVerify: options.Insecure},
	}
	if options.EnrollmentToken != "" {
		enrollment, err := Enroll(result.Master.APIURL, options.EnrollmentToken, options.Insecure)
		if err != nil {
			return Config{}, fmt.Errorf("enroll server: %w", err)
		}
		result.Credentials.ServerID = enrollment.ServerID
		result.Credentials.Secret = enrollment.Secret
	}
	if err := result.Validate(); err != nil {
		return Config{}, fmt.Errorf("server is not enrolled and no valid persisted configuration exists: %w", err)
	}
	if err := WriteConfig(options.ConfigPath, result); err != nil {
		return Config{}, err
	}
	return result, nil
}

func Enroll(apiURL, token string, insecureSkipVerify bool) (EnrollmentResult, error) {
	base, err := url.Parse(apiURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return EnrollmentResult{}, fmt.Errorf("invalid Master API URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/api/v2/server/enroll"
	base.RawQuery = ""
	base.Fragment = ""
	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return EnrollmentResult{}, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecureSkipVerify} // #nosec G402 -- explicit operator opt-in
	client := &http.Client{Timeout: 20 * time.Second, Transport: transport}
	request, err := http.NewRequest(http.MethodPost, base.String(), bytes.NewReader(body))
	if err != nil {
		return EnrollmentResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return EnrollmentResult{}, fmt.Errorf("redeem enrollment: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return EnrollmentResult{}, fmt.Errorf("read enrollment response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		var problem struct {
			Detail string `json:"detail"`
		}
		_ = json.Unmarshal(raw, &problem)
		if problem.Detail == "" {
			problem.Detail = response.Status
		}
		return EnrollmentResult{}, fmt.Errorf("redeem enrollment: %s", problem.Detail)
	}
	var result EnrollmentResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return EnrollmentResult{}, fmt.Errorf("decode enrollment response: %w", err)
	}
	if result.ServerID == "" || result.Secret == "" {
		return EnrollmentResult{}, fmt.Errorf("Master returned incomplete enrollment credentials")
	}
	return result, nil
}

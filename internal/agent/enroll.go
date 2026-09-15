package agent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type EnrollmentResult struct {
	NodeID string `json:"nodeId"`
	Secret string `json:"secret"`
}

func Enroll(apiURL, token string, insecureSkipVerify bool) (EnrollmentResult, error) {
	base, err := url.Parse(apiURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return EnrollmentResult{}, fmt.Errorf("invalid controller API URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/api/v2/agent/enroll"
	base.RawQuery = ""
	base.Fragment = ""
	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return EnrollmentResult{}, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecureSkipVerify} // #nosec G402 -- explicit CLI opt-in
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
	if result.NodeID == "" || result.Secret == "" {
		return EnrollmentResult{}, fmt.Errorf("controller returned incomplete enrollment credentials")
	}
	return result, nil
}

package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type TokenDetails struct {
	Token  string
	Expiry time.Duration
}

func GetTokenFromF5(url, username, password string, expiry time.Duration, insecure bool, timeout time.Duration) (TokenDetails, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: timeout * time.Second}
	loginURL := url + "/api/data/openconfig-system:system/aaa"

	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return TokenDetails{}, fmt.Errorf("error creating login request: %w", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/yang-data+json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenDetails{}, fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenDetails{}, fmt.Errorf("error reading login response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return TokenDetails{}, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("New authentication token obtained from %s (expires in %s)", url, expiry)

	return TokenDetails{Token: resp.Header.Get("X-Auth-Token"), Expiry: expiry}, nil
}

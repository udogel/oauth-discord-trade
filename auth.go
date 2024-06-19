package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ExchangeResponse struct {
	Type         string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	Expires      int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func AuthenticateUser(config Config) (ExchangeResponse, error) {
	var exchangeResponse ExchangeResponse

	code, err := getCode(config)
	if err != nil {
		return exchangeResponse, fmt.Errorf("failed to get code: %w", err)
	}

	exchangeResponse, err = getAuth(config, code)
	if err != nil {
		return exchangeResponse, fmt.Errorf("failed to exchange code: %w", err)
	}

	return exchangeResponse, nil
}

func getCode(config Config) (string, error) {
	exchangeCode := make(chan string)
	shutdownCh := make(chan error)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/favicon.io" {
			code := r.URL.String()[7:37]
			exchangeCode <- code
			fmt.Fprint(w, "Code received, you can now close this window.")
		}
	})

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			shutdownCh <- err
		}
	}()

	select {
	case exchangeCode := <-exchangeCode:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return "", fmt.Errorf("failed to shutdown server: %w", err)
		}
		return exchangeCode, nil
	case err := <-shutdownCh:
		return "", fmt.Errorf("server shutdown unexpectedly: %w", err)
	case <-time.After(time.Duration(config.TimeoutTime) * time.Second):
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		return "", fmt.Errorf("timed out waiting for exchange code")
	}
}

func getAuth(config Config, code string) (ExchangeResponse, error) {
	var exchangeResponse ExchangeResponse

	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code", code)
	params.Set("redirect_uri", "http://localhost:"+config.Port)
	params.Set("client_id", config.ClientID)
	params.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", strings.NewReader(params.Encode()))
	if err != nil {
		return exchangeResponse, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return exchangeResponse, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return exchangeResponse, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &exchangeResponse); err != nil {
			return exchangeResponse, fmt.Errorf("failed to parse response: %w", err)
		}
	} else {
		return exchangeResponse, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
	}

	return exchangeResponse, nil
}

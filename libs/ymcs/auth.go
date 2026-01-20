/*
 * Copyright (C) 2026 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * This file is part of Falconieri project.
 *
 * Falconieri is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * or any later version.
 *
 * Falconieri is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Falconieri.  If not, see LICENSE.
 *
 * author: Matteo Valentini <matteo.valentini@nethesis.it>
 */

// Package ymcs provides a client for the Yealink YMCS API with OAuth 2.0 authentication
package ymcs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GetAccessToken retrieves and caches an OAuth 2.0 access token.
// It uses the client credentials flow and automatically handles token caching
// and refresh when the token expires.
func (c *Client) GetAccessToken() (string, error) {
	// serialized access to token cache
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return cached token if still valid
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		if c.Debug {
			log.Printf("[YMCS] Using cached access token")
		}
		return c.accessToken, nil
	}

	if c.Debug {
		log.Printf("[YMCS] Requesting new access token from %s/v2/token", c.BaseURL)
	}

	// Create Basic Auth credentials
	credentials := fmt.Sprintf("%s:%s", c.ClientID, c.ClientSecret)
	encodedCreds := base64.StdEncoding.EncodeToString([]byte(credentials))

	// Prepare request
	requestBody := map[string]string{
		"grant_type": "client_credentials",
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v2/token", c.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	nonce, err := c.getNonce()
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+encodedCreds)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("timestamp", c.getTimestamp())
	req.Header.Set("nonce", nonce)

	// Store request for debugging
	c.debugMu.Lock()
	c.LastRequest = req
	c.LastRequestBody = string(jsonData)
	c.debugMu.Unlock()

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Store response for debugging
	c.debugMu.Lock()
	c.LastResponse = resp
	c.debugMu.Unlock()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	c.debugMu.Lock()
	c.LastRespBody = string(respBody)
	c.debugMu.Unlock()

	if resp.StatusCode != http.StatusOK {
		return "", parseAPIError(resp.StatusCode, respBody)
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache token
	c.accessToken = tokenResp.AccessToken
	// Set expiry with 60 second buffer
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return c.accessToken, nil
}

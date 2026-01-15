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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents a Yealink YMCS API client
type Client struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	Debug        bool // Enable debug logging

	// Token management
	accessToken string
	tokenExpiry time.Time

	// Last request/response for debugging
	LastRequest     *http.Request
	LastRequestBody string
	LastResponse    *http.Response
	LastRespBody    string
}

// NewClient creates a new Yealink YMCS API client
func NewClient(baseURL, clientID, clientSecret string) *Client {
	return &Client{
		BaseURL:      strings.TrimSuffix(baseURL, "/"),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// getTimestamp returns current timestamp in milliseconds
func (c *Client) getTimestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

// getNonce generates a random 32-character hex string
func (c *Client) getNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// makeRequest makes an authenticated API request
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	// Get valid token
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare request body
	var reqBody io.Reader
	var requestBodyStr string
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBodyStr = string(jsonData)
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Create request
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Store request for debugging
	c.LastRequest = req
	c.LastRequestBody = requestBodyStr

	// Add headers
	nonce, err := c.getNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("timestamp", c.getTimestamp())
	req.Header.Set("nonce", nonce)

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Store response for debugging
	c.LastResponse = resp

	// Read response body for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	c.LastRespBody = string(bodyBytes)

	// Recreate response body reader
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return resp, nil
}

// cleanMAC removes separators from MAC address and converts to lowercase
func cleanMAC(mac string) string {
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	return strings.ToLower(mac)
}

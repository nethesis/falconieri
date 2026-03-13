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

package grape

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HawkCredentials represents the Hawk authentication credentials
type HawkCredentials struct {
	ID  string
	Key string
}

// generateNonce creates a cryptographically secure nonce
func generateNonce() string {
	// Generate 16 bytes (128 bits) of random data
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to nanosecond timestamp if crypto/rand fails
		// This is extremely rare but provides a safety net
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	// Return base64-encoded random bytes (URL-safe encoding)
	return base64.URLEncoding.EncodeToString(bytes)
}

// calculatePayloadHash computes SHA256 hash of the payload
func calculatePayloadHash(payload []byte, contentType string) string {
	// Hawk payload hash format: hawk.1.payload\n{content-type}\n{payload}\n
	hashContent := fmt.Sprintf("hawk.1.payload\n%s\n", contentType)
	if payload != nil {
		hashContent += string(payload)
	}
	hashContent += "\n"

	hash := sha256.Sum256([]byte(hashContent))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// calculateMAC computes the HMAC-SHA256 for Hawk authentication
func calculateMAC(creds *HawkCredentials, method, uri, host, port, timestamp, nonce, payloadHash string) string {
	// Construct the normalized request string according to Hawk specification
	normalized := fmt.Sprintf("hawk.1.header\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n\n",
		timestamp,
		nonce,
		strings.ToUpper(method),
		uri,
		host,
		port,
		payloadHash)

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(creds.Key))
	h.Write([]byte(normalized))
	mac := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(mac)
}

// createHawkAuthHeader creates the Hawk Authorization header
func createHawkAuthHeader(creds *HawkCredentials, method, rawURL string, payload []byte, contentType string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	// Extract components needed for Hawk
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// Generate timestamp and nonce
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce()

	// Calculate payload hash
	var payloadHash string
	if payload != nil && len(payload) > 0 {
		payloadHash = calculatePayloadHash(payload, contentType)
	} else {
		payloadHash = calculatePayloadHash(nil, "")
	}

	// Calculate MAC
	uri := parsedURL.RequestURI()
	mac := calculateMAC(creds, method, uri, host, port, timestamp, nonce, payloadHash)

	// Construct the Authorization header
	authHeader := fmt.Sprintf(`Hawk id="%s", ts="%s", nonce="%s", hash="%s", mac="%s"`,
		creds.ID, timestamp, nonce, payloadHash, mac)

	return authHeader, nil
}

// makeHawkRequestFull performs an HTTP request with Hawk authentication and
// returns both the response body and headers (needed for Link-based pagination).
func (c *Client) makeHawkRequestFull(method, url string, body []byte) ([]byte, http.Header, error) {
	var req *http.Request
	var err error
	var contentType string

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			return nil, nil, err
		}
		contentType = "application/json"
		req.Header.Set("Content-Type", contentType)
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, nil, err
		}
		contentType = ""
	}

	req.Header.Set("Accept", "application/json")

	// Create Hawk credentials
	creds := &HawkCredentials{
		ID:  c.ClientID,
		Key: c.ClientSecret,
	}

	// Add Hawk authentication header
	authHeader, err := createHawkAuthHeader(creds, method, url, body, contentType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Hawk auth header: %v", err)
	}

	req.Header.Set("Authorization", authHeader)

	// Store request for debugging
	if c.Debug {
		c.debugMu.Lock()
		c.LastRequest = req
		c.LastRequestBody = string(body)
		c.debugMu.Unlock()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Store response for debugging
	if c.Debug {
		c.debugMu.Lock()
		c.LastResponse = resp
		c.LastRespBody = string(bodyBytes)
		c.debugMu.Unlock()
	}

	if resp.StatusCode >= 400 {
		return nil, nil, parseErrorResponse(resp.StatusCode, resp.Status, bodyBytes)
	}

	return bodyBytes, resp.Header, nil
}

// makeHawkRequest performs an HTTP request with Hawk authentication
func (c *Client) makeHawkRequest(method, url string, body []byte) ([]byte, error) {
	bodyBytes, _, err := c.makeHawkRequestFull(method, url, body)
	return bodyBytes, err
}

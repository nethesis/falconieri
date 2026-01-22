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
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Client represents a Grape API client
type Client struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	Debug        bool

	// Cached API navigation data (thread-safe)
	mu                   sync.Mutex
	provisioningServerID string
	endpointsURL         string

	// Debug fields
	LastRequest     *http.Request
	LastRequestBody string
	LastResponse    *http.Response
	LastRespBody    string
}

// NewClient creates a new Grape API client
func NewClient(baseURL, clientID, clientSecret string) *Client {
	return &Client{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Debug: false,
	}
}

// normalizeMac converts any MAC address format to lowercase without separators
// Accepts formats: AA:BB:CC:DD:EE:FF, AA-BB-CC-DD-EE-FF, AABBCCDDEEFF, aabbccddeeff
// Returns: aabbccddeeff
func normalizeMac(mac string) string {
	// Remove common separators: colons, hyphens, dots, spaces
	re := regexp.MustCompile(`[:\-\.\s]`)
	normalized := re.ReplaceAllString(mac, "")
	// Convert to lowercase
	return strings.ToLower(normalized)
}

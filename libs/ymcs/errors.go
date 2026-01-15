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

package ymcs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIErrorDetail represents a single validation error returned by YMCS.
type APIErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError models the standard YMCS error response payload and keeps HTTP context.
type APIError struct {
	HTTPStatus      int              `json:"-"`
	Code            string           `json:"code"`
	OAuthError      string           `json:"error"`
	RequestID       string           `json:"requestId"`
	Message         string           `json:"message"`
	Details         []APIErrorDetail `json:"details"`
	RawBody         string           `json:"-"`
	FriendlyMessage string           `json:"-"`
}

// Error implements the error interface and exposes a compact summary.
func (e APIError) Error() string {
	parts := []string{}

	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("code=%s", e.Code))
	}
	if e.OAuthError != "" {
		parts = append(parts, fmt.Sprintf("oauth=%s", e.OAuthError))
	}
	if e.RequestID != "" {
		parts = append(parts, fmt.Sprintf("requestId=%s", e.RequestID))
	}

	// Prefer the descriptive message if present.
	msg := e.Message
	if msg == "" {
		msg = e.FriendlyMessage
	}
	if msg != "" {
		parts = append(parts, fmt.Sprintf("message=%s", msg))
	}

	if e.FriendlyMessage != "" && e.FriendlyMessage != msg {
		parts = append(parts, fmt.Sprintf("hint=%s", e.FriendlyMessage))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("ymcs error (status %d)", e.HTTPStatus)
	}

	return fmt.Sprintf("ymcs error (status %d): %s", e.HTTPStatus, strings.Join(parts, ", "))
}

// knownErrorCodes captures the main YMCS error codes documented in the API guide.
var knownErrorCodes = map[string]string{
	"70011":  "Invalid grant_type", // OAuth example
	"800001": "Invalid MAC",
	"800002": "Invalid SN",
	"800003": "Resource already exists",
	"800004": "Device already managed by another organization",
	"800005": "Device type is invalid",
	"800006": "Batch size exceeds limit",
	"800007": "Invalid parameters",
	"800008": "Data exceeds limit",
	"800130": "Account already exists",
	"800200": "Device type is invalid",
	"900200": "Operation successful", // Included for completeness
	"900400": "Request parameters are incorrect or missing",
	"900401": "Not logged in or session expired",
	"900403": "Request forbidden",
	"900404": "Resource not found",
	"900408": "Request timed out",
	"900409": "Request conflict",
	"900412": "Concurrent editing error",
	"900429": "Too many requests (rate limited)",
	"900440": "Session expired",
	"900500": "Server busy, try again later",
	"900501": "Operation not supported",
	"900502": "Bad gateway",
	"900503": "Service unavailable",
	"900504": "Gateway timeout",
	"900511": "Internal server error",
	"900599": "Unknown server error",
}

// parseAPIError tries to decode the YMCS error payload, enriching it with friendly messages.
// When the payload cannot be decoded, it falls back to a generic error that still carries the status code.
func parseAPIError(status int, body []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		// Fall back to a generic error with the raw payload to preserve context.
		return fmt.Errorf("ymcs request failed with status %d: %s", status, strings.TrimSpace(string(body)))
	}

	apiErr.HTTPStatus = status
	apiErr.RawBody = string(body)
	apiErr.FriendlyMessage = knownErrorCodes[apiErr.Code]

	// Fill missing message fields with whatever context we have.
	if apiErr.Message == "" {
		apiErr.Message = apiErr.FriendlyMessage
	}
	if apiErr.Message == "" && apiErr.OAuthError != "" {
		apiErr.Message = apiErr.OAuthError
	}

	return apiErr
}

// statusAllowed reports whether the HTTP status code is one of the expected values.
func statusAllowed(status int, allowed []int) bool {
	for _, s := range allowed {
		if status == s {
			return true
		}
	}
	return false
}

// decodeResponse validates the status code, decodes the body into the provided destination and returns
// a structured APIError when the response contains a YMCS error payload.
func decodeResponse(resp *http.Response, dst interface{}, allowedStatuses ...int) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if !statusAllowed(resp.StatusCode, allowedStatuses) {
		return parseAPIError(resp.StatusCode, body)
	}

	if dst == nil || len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

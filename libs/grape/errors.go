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
	"encoding/json"
	"fmt"
)

// APIError represents an error response from the GRAPE API
type APIError struct {
	StatusCode int
	Status     string
	Message    string
	Body       string
}

// Error implements the error interface
func (e APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("GRAPE API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("GRAPE API error (HTTP %d): %s", e.StatusCode, e.Status)
}

// parseErrorResponse attempts to extract error information from the response body
func parseErrorResponse(statusCode int, status string, body []byte) error {
	apiErr := APIError{
		StatusCode: statusCode,
		Status:     status,
		Body:       string(body),
	}

	// Try to parse JSON error response
	var errResp map[string]interface{}
	if err := json.Unmarshal(body, &errResp); err == nil {
		// Extract error message if available
		if msg, ok := errResp["message"].(string); ok {
			apiErr.Message = msg
		} else if msg, ok := errResp["error"].(string); ok {
			apiErr.Message = msg
		} else if detail, ok := errResp["detail"].(string); ok {
			apiErr.Message = detail
		}
	}

	// If no message was extracted, use the raw body
	if apiErr.Message == "" && len(body) > 0 {
		apiErr.Message = string(body)
	}

	return apiErr
}

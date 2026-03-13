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
	"strings"
)

// getProvisioningServerID retrieves and caches the provisioning setting UUID.
// The setting name is determined by c.ProvisioningSettingName (e.g.
// "ProvisioningServer" for Grape, "setting_server" for SRAPS).
// Uses sync.Once to prevent duplicate API calls under concurrent load.
func (c *Client) getProvisioningServerID() (string, error) {
	c.provisioningServerIDOnce.Do(func() {
		// Fetch all settings pages to find the provisioning setting UUID.
		// The API paginates with default 20 items; use page_size=1000 to
		// minimise round-trips. Pagination continues while the response
		// contains a Link header with rel="next".
		nextURL := c.BaseURL + "settings/?page_size=1000"
		for nextURL != "" {
			pageBytes, headers, err := c.makeHawkRequestFull("GET", nextURL, nil)
			if err != nil {
				c.provisioningServerIDErr = fmt.Errorf("failed to get settings: %w", err)
				return
			}

			var page []Setting
			if err := json.Unmarshal(pageBytes, &page); err != nil {
				c.provisioningServerIDErr = fmt.Errorf("failed to parse settings response: %w", err)
				return
			}

			for _, setting := range page {
				if setting.ParamName == c.ProvisioningSettingName {
					c.provisioningServerID = setting.UUID
					return
				}
			}

			// Follow Link header for next page if present
			nextURL = parseLinkNext(headers.Get("Link"))
		}

		c.provisioningServerIDErr = fmt.Errorf("%s setting not found in API response", c.ProvisioningSettingName)
	})

	return c.provisioningServerID, c.provisioningServerIDErr
}

// parseLinkNext extracts the URL for rel="next" from an RFC 5988 Link header.
// Returns an empty string if there is no next page.
func parseLinkNext(linkHeader string) string {
	// Format: <https://...>; rel="next", <https://...>; rel="prev"
	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		segments := strings.Split(part, ";")
		if len(segments) < 2 {
			continue
		}
		for _, seg := range segments[1:] {
			if strings.TrimSpace(seg) == `rel="next"` {
				url := strings.TrimSpace(segments[0])
				url = strings.TrimPrefix(url, "<")
				url = strings.TrimSuffix(url, ">")
				return url
			}
		}
	}
	return ""
}

// getEndpointsURL retrieves and caches the endpoints URL for device operations
// Uses sync.Once to prevent duplicate API calls under concurrent load
func (c *Client) getEndpointsURL() (string, error) {
	c.endpointsURLOnce.Do(func() {
		// Get company endpoints URL
		tokenURL := c.BaseURL + "tokens/" + c.ClientID
		tokenResp, err := c.makeHawkRequest("GET", tokenURL, nil)
		if err != nil {
			c.endpointsURLErr = fmt.Errorf("failed to get token info: %w", err)
			return
		}

		var tokenData TokenResponse
		if err := json.Unmarshal(tokenResp, &tokenData); err != nil {
			c.endpointsURLErr = fmt.Errorf("failed to parse token response: %w", err)
			return
		}

		companyResp, err := c.makeHawkRequest("GET", tokenData.Links.Company, nil)
		if err != nil {
			c.endpointsURLErr = fmt.Errorf("failed to get company info: %w", err)
			return
		}

		var companyData CompanyResponse
		if err := json.Unmarshal(companyResp, &companyData); err != nil {
			c.endpointsURLErr = fmt.Errorf("failed to parse company response: %w", err)
			return
		}

		c.endpointsURL = companyData.Links.Endpoints
	})

	return c.endpointsURL, c.endpointsURLErr
}

// RegisterDevice registers a device with the Grape provisioning server
// mac: MAC address in any format (will be normalized)
// provisioningURL: The URL of the provisioning server
func (c *Client) RegisterDevice(mac, provisioningURL string) error {
	// Normalize MAC address
	normalizedMAC := normalizeMac(mac)

	// Get ProvisioningServer UUID (cached)
	settingProvisioningServerUUID, err := c.getProvisioningServerID()
	if err != nil {
		return err
	}

	// Get endpoints URL (cached)
	endpointsURL, err := c.getEndpointsURL()
	if err != nil {
		return err
	}

	// Add device
	deviceData := DeviceData{
		MAC:                        normalizedMAC,
		AutoprovisioningEnabled:    true,
		WarrantyExpWarningPeriod:   nil,
		SettingsManager: map[string]map[string]interface{}{
			settingProvisioningServerUUID: {
				"value": provisioningURL,
				"attrs": map[string]string{"perm": "RW"},
			},
		},
	}

	deviceJSON, err := json.Marshal(deviceData)
	if err != nil {
		return fmt.Errorf("failed to marshal device data: %w", err)
	}

	deviceURL := endpointsURL + normalizedMAC
	_, err = c.makeHawkRequest("PUT", deviceURL, deviceJSON)
	if err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}

	return nil
}

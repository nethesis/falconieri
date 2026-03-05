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

package sraps

import (
	"encoding/json"
	"fmt"
)

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

// RegisterDevice registers a device with the SRAPS provisioning server
// mac: MAC address in any format (will be normalized)
// provisioningURL: The URL of the provisioning server
func (c *Client) RegisterDevice(mac, provisioningURL string) error {
	// Normalize MAC address
	normalizedMAC := normalizeMac(mac)

	// Get endpoints URL (cached)
	endpointsURL, err := c.getEndpointsURL()
	if err != nil {
		return err
	}

	// SRAPS uses a simpler device structure without ProvisioningServer UUID
	// The provisioning URL is set directly in the device data
	deviceData := DeviceData{
		MAC:                     normalizedMAC,
		AutoprovisioningEnabled: true,
		ProvisioningURL:         provisioningURL,
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

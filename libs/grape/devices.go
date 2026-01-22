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
	"errors"
	"fmt"
)

// getProvisioningServerID retrieves and caches the ProvisioningServer UUID
func (c *Client) getProvisioningServerID() (string, error) {
	c.mu.Lock()
	if c.provisioningServerID != "" {
		id := c.provisioningServerID
		c.mu.Unlock()
		return id, nil
	}
	c.mu.Unlock()

	// Get settings to find ProvisioningServer UUID
	settingsURL := c.BaseURL + "settings/"
	settings, err := c.makeHawkRequest("GET", settingsURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %w", err)
	}

	var settingsList []Setting
	if err := json.Unmarshal(settings, &settingsList); err != nil {
		return "", fmt.Errorf("failed to parse settings response: %w", err)
	}

	var settingProvisioningServerUUID string
	for _, setting := range settingsList {
		if setting.ParamName == "ProvisioningServer" {
			settingProvisioningServerUUID = setting.UUID
			break
		}
	}

	if settingProvisioningServerUUID == "" {
		return "", errors.New("ProvisioningServer setting not found in API response")
	}

	// Cache the result
	c.mu.Lock()
	c.provisioningServerID = settingProvisioningServerUUID
	c.mu.Unlock()

	return settingProvisioningServerUUID, nil
}

// getEndpointsURL retrieves and caches the endpoints URL for device operations
func (c *Client) getEndpointsURL() (string, error) {
	c.mu.Lock()
	if c.endpointsURL != "" {
		url := c.endpointsURL
		c.mu.Unlock()
		return url, nil
	}
	c.mu.Unlock()

	// Get company endpoints URL
	tokenURL := c.BaseURL + "tokens/" + c.ClientID
	tokenResp, err := c.makeHawkRequest("GET", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get token info: %w", err)
	}

	var tokenData TokenResponse
	if err := json.Unmarshal(tokenResp, &tokenData); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	companyResp, err := c.makeHawkRequest("GET", tokenData.Links.Company, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get company info: %w", err)
	}

	var companyData CompanyResponse
	if err := json.Unmarshal(companyResp, &companyData); err != nil {
		return "", fmt.Errorf("failed to parse company response: %w", err)
	}

	// Cache the result
	c.mu.Lock()
	c.endpointsURL = companyData.Links.Endpoints
	c.mu.Unlock()

	return companyData.Links.Endpoints, nil
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

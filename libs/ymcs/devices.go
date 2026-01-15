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
	"fmt"
	"net/http"
)

// SearchDevices searches for devices with optional MAC filter.
// Use skip and limit for pagination, and set autoCount to true to get total count.
func (c *Client) SearchDevices(mac string, skip, limit int, autoCount bool) (*DeviceSearchResponse, error) {
	req := DeviceSearchRequest{
		Skip:      skip,
		Limit:     limit,
		AutoCount: autoCount,
	}

	// Add MAC filter if provided
	if mac != "" {
		req.Filter = &DeviceFilter{
			MAC: cleanMAC(mac),
		}
	}

	resp, err := c.makeRequest("POST", "/v2/rps/listDevices", req)
	if err != nil {
		return nil, err
	}

	var result DeviceSearchResponse
	if err := decodeResponse(resp, &result, http.StatusOK); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDeviceDetails retrieves detailed information about a device by its ID.
func (c *Client) GetDeviceDetails(deviceID string) (*DeviceDetails, error) {
	endpoint := fmt.Sprintf("/v2/rps/devices/%s", deviceID)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var details DeviceDetails
	if err := decodeResponse(resp, &details, http.StatusOK); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetDevicePINs retrieves PINs for multiple devices.
// MAC addresses can be provided in any format (with or without separators).
func (c *Client) GetDevicePINs(macs []string) ([]DevicePIN, error) {
	// Clean MAC addresses
	cleanedMACs := make([]string, len(macs))
	for i, mac := range macs {
		cleanedMACs[i] = cleanMAC(mac)
	}

	req := DevicePINRequest{
		MACs: cleanedMACs,
	}

	resp, err := c.makeRequest("POST", "/v2/rps/listDevicePins", req)
	if err != nil {
		return nil, err
	}

	var pins []DevicePIN
	if err := decodeResponse(resp, &pins, http.StatusOK); err != nil {
		return nil, err
	}

	return pins, nil
}

// GetSingleDevicePIN retrieves PIN for a single device.
// This is a convenience method that calls GetDevicePINs with a single MAC address.
func (c *Client) GetSingleDevicePIN(mac string) (string, error) {
	pins, err := c.GetDevicePINs([]string{mac})
	if err != nil {
		return "", err
	}

	if len(pins) == 0 {
		// PIN is optional; surface empty value without failing the call.
		return "", nil
	}

	return pins[0].PIN, nil
}

// AddDevice adds a new device to the YMCS system.
// Both MAC address and serial number (SN) are required.
func (c *Client) AddDevice(req AddDeviceRequest) (*AddDeviceResponse, error) {
	// Clean MAC address
	req.MAC = cleanMAC(req.MAC)

	resp, err := c.makeRequest("POST", "/v2/rps/devices", req)
	if err != nil {
		return nil, err
	}

	var result AddDeviceResponse
	if err := decodeResponse(resp, &result, http.StatusCreated); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddDevicesByMac adds one or more devices without serial numbers (batch operation).
// Maximum of 100 devices can be added in a single request.
func (c *Client) AddDevicesByMac(devices []AddDeviceByMacRequest) (*AddDevicesByMacResponse, error) {
	// Clean MAC addresses
	for i := range devices {
		devices[i].MAC = cleanMAC(devices[i].MAC)
	}

	resp, err := c.makeRequest("POST", "/v2/rps/addDevicesByMac", devices)
	if err != nil {
		return nil, err
	}

	var result AddDevicesByMacResponse
	if err := decodeResponse(resp, &result, http.StatusOK); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddDeviceByMacSingle is a convenience method to add a single device by MAC without SN.
func (c *Client) AddDeviceByMacSingle(mac string, uniqueServerURL *string) (*AddDevicesByMacResponse, error) {
	req := AddDeviceByMacRequest{
		MAC:             mac,
		UniqueServerURL: uniqueServerURL,
	}
	return c.AddDevicesByMac([]AddDeviceByMacRequest{req})
}

// DeleteDevices deletes one or more devices from the YMCS system.
// deviceIdType must be either "mac" or "id".
func (c *Client) DeleteDevices(deviceIdType string, deviceIds []string) (*DeleteDevicesResponse, error) {
	// Clean device IDs if using MAC addresses
	if deviceIdType == "mac" {
		cleanedIds := make([]string, len(deviceIds))
		for i, id := range deviceIds {
			cleanedIds[i] = cleanMAC(id)
		}
		deviceIds = cleanedIds
	}

	req := DeleteDevicesRequest{
		DeviceIdType: deviceIdType,
		DeviceIds:    deviceIds,
	}

	resp, err := c.makeRequest("POST", "/v2/rps/delDevices", req)
	if err != nil {
		return nil, err
	}

	var result DeleteDevicesResponse
	if err := decodeResponse(resp, &result, http.StatusOK); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteDeviceByMAC is a convenience method to delete a single device by MAC address.
func (c *Client) DeleteDeviceByMAC(mac string) (*DeleteDevicesResponse, error) {
	return c.DeleteDevices("mac", []string{mac})
}

// DeleteDeviceByID is a convenience method to delete a single device by ID.
func (c *Client) DeleteDeviceByID(deviceID string) (*DeleteDevicesResponse, error) {
	return c.DeleteDevices("id", []string{deviceID})
}

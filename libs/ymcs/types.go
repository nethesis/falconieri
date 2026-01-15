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

// TokenResponse represents the OAuth 2.0 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Device represents a device in the YMCS system
type Device struct {
	ID              string  `json:"id"`
	MAC             string  `json:"mac"`
	SN              *string `json:"sn"`
	ServerID        *string `json:"serverId"`
	ServerName      *string `json:"serverName"`
	ServerURL       *string `json:"serverUrl"`
	UniqueServerURL *string `json:"uniqueServerUrl"`
	IPAddress       *string `json:"ipAddress"`
	Remark          *string `json:"remark"`
	DateRegistered  *int64  `json:"dateRegistered"`
	LastConnected   *int64  `json:"lastConnected"`
}

// DeviceDetails represents detailed device information
type DeviceDetails struct {
	Device
	AuthName *string `json:"authName"`
}

// DeviceSearchRequest represents the request body for device search
type DeviceSearchRequest struct {
	Skip      int           `json:"skip"`
	Limit     int           `json:"limit"`
	AutoCount bool          `json:"autoCount"`
	Filter    *DeviceFilter `json:"filter,omitempty"`
}

// DeviceFilter represents search filters
type DeviceFilter struct {
	MAC string `json:"mac,omitempty"`
}

// DeviceSearchResponse represents the response from device search
type DeviceSearchResponse struct {
	Skip  int      `json:"skip"`
	Limit int      `json:"limit"`
	Total int      `json:"total"`
	Data  []Device `json:"data"`
}

// DevicePIN represents a device PIN
type DevicePIN struct {
	MAC string `json:"mac"`
	PIN string `json:"pin"`
}

// DevicePINRequest represents the request for device PINs
type DevicePINRequest struct {
	MACs []string `json:"macs"`
}

// AddDeviceRequest represents the request body for adding a device
type AddDeviceRequest struct {
	MAC             string  `json:"mac"`
	SN              string  `json:"sn"`
	ServerID        *string `json:"serverId,omitempty"`
	UniqueServerURL *string `json:"uniqueServerUrl,omitempty"`
	AuthName        *string `json:"authName,omitempty"`
	Password        *string `json:"password,omitempty"`
	Remark          *string `json:"remark,omitempty"`
}

// AddDeviceResponse represents the response from adding a device
type AddDeviceResponse struct {
	ID              string  `json:"id"`
	MAC             string  `json:"mac"`
	SN              string  `json:"sn"`
	ServerID        *string `json:"serverId,omitempty"`
	UniqueServerURL *string `json:"uniqueServerUrl,omitempty"`
	AuthName        *string `json:"authName,omitempty"`
	Remark          *string `json:"remark,omitempty"`
}

// AddDeviceByMacRequest represents a single device in batch add by MAC (no SN required)
type AddDeviceByMacRequest struct {
	MAC             string  `json:"mac"`
	ServerID        *string `json:"serverId,omitempty"`
	UniqueServerURL *string `json:"uniqueServerUrl,omitempty"`
	AuthName        *string `json:"authName,omitempty"`
	Password        *string `json:"password,omitempty"`
	Remark          *string `json:"remark,omitempty"`
}

// AddError represents an error when adding a device
type AddError struct {
	MAC       string `json:"mac"`
	SN        string `json:"sn"`
	ErrorInfo string `json:"errorInfo"`
}

// AddDevicesByMacResponse represents the response from batch adding devices by MAC
type AddDevicesByMacResponse struct {
	Total        int        `json:"total"`
	SuccessCount int        `json:"successCount"`
	FailureCount int        `json:"failureCount"`
	Errors       []AddError `json:"errors"`
}

// DeleteDevicesRequest represents the request body for deleting devices
type DeleteDevicesRequest struct {
	DeviceIdType string   `json:"deviceIdType"` // "mac" or "id"
	DeviceIds    []string `json:"deviceIds"`    // List of device IDs or MACs
}

// OpError represents an error for a specific field
type OpError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

// DeleteDevicesResponse represents the response from deleting devices
type DeleteDevicesResponse struct {
	Total        int       `json:"total"`
	SuccessCount int       `json:"successCount"`
	FailureCount int       `json:"failureCount"`
	Errors       []OpError `json:"errors"`
}

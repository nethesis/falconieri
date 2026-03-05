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

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	Links struct {
		Company string `json:"company"`
	} `json:"links"`
}

// CompanyResponse represents the response from the company endpoint
type CompanyResponse struct {
	Links struct {
		Endpoints string `json:"endpoints"`
	} `json:"links"`
}

// DeviceData represents the data structure for registering a device
// SRAPS uses a simpler structure than GRAPE
type DeviceData struct {
	MAC                     string `json:"mac"`
	AutoprovisioningEnabled bool   `json:"autoprovisioning_enabled"`
	ProvisioningURL         string `json:"provisioning_url"`
}

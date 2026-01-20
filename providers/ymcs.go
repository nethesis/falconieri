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

/*
 * Yealink Management Cloud Service provider
 */
package providers

import (
	"errors"
	"sync"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/libs/ymcs"
	"github.com/nethesis/falconieri/models"
	"github.com/nethesis/falconieri/utils"
)

var (
	ymcsClient     *ymcs.Client
	ymcsClientOnce sync.Once
)

// getYmcsClient returns a singleton YMCS client instance.
// This allows token caching to work effectively across multiple requests.
// The client is created using configuration loaded at startup.
// Errors during API calls are handled by the client's methods.
func getYmcsClient() *ymcs.Client {
	ymcsClientOnce.Do(func() {
		ymcsClient = ymcs.NewClient(
			configuration.Config.Providers.Ymcs.BaseURL,
			configuration.Config.Providers.Ymcs.ClientID,
			configuration.Config.Providers.Ymcs.ClientSecret,
		)
	})
	return ymcsClient
}

type YmcsDevice struct {
	Mac string
	Url string
}

func (s YmcsDevice) Register() error {
	client := getYmcsClient()

	// Delete device first (if it exists) to ensure clean registration
	_, err := client.DeleteDeviceByMAC(s.Mac)
	if err != nil {
		// Ignore deletion errors (device may not exist)
		// Log but continue with registration
	}

	resp, err := client.AddDeviceByMacSingle(s.Mac, &s.Url)
	if err != nil {
		return models.ProviderError{
			Message:      "provider_remote_call_failed",
			WrappedError: err,
		}
	}

	if resp == nil {
		return errors.New("unknown_response_from_provider")
	}

	if resp.FailureCount > 0 && len(resp.Errors) > 0 {
		return utils.ParseProviderError(resp.Errors[0].ErrorInfo)
	}

	if resp.SuccessCount == 0 {
		return errors.New("provider_remote_call_failed")
	}

	return nil
}

func (s YmcsDevice) GetPIN() (string, error) {
	client := getYmcsClient()

	return client.GetSingleDevicePIN(s.Mac)
}

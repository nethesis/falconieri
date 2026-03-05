# Falconieri

Modern IP phones can contact a redirect service provider at boot time
and discover their PBX address with it.

The Falconieri project is a RPS (Redirect and Provisioning Service) gateway 
that helps to store the phone provisioning URL in the phone vendor redirect service.

Supported providers:

* [Snom](https://service.snom.com/display/wiki/XML-RPC+API)
* [Gigaset](https://teamwork.gigaset.com/gigawiki/display/GPPPO/Gigaset+Redirect+server)
* Fanvil (a link to its public documentation was not found)
* [Yealink (legacy provider)](https://support-cdn.yealink.com/attachment/upload/attachment/2019-1-8/5/b6a08cc4-0d6c-4def-b2f4-224b9653c051/Yealink+XML+API+for+RPS-V1.6-ENG.pdf)
* [YMCS (Yealink Management Cloud Service V4X)](https://support.yealink.com/document-detail/c0966bbacb51405397c55290c2925f65) To use the YMCS provider, you need to ask Yealink to enable `/v2/rps/addDevicesByMac` endpoint for your account.
* [Grape (Gigaset Redirect and Provisioning Environment)](https://teamwork.gigaset.com/gigawiki/pages/viewpage.action?pageId=1535868981)
* SRAPS (Secure Redirection and Provisioning Service) - SNOM provider using the same protocol as GRAPE

## APIs

### PUT /providers/:provider/:mac
---

Register device on remote provisioning service, if the device is already configured,
the new configuration override the old.

##### Request

###### Headers
* `Content-Type: application/json`

###### Path variables
* `provider`: Name of the remote provider.
* `mac`: Mac address of the device, represented in the EUI-48 IEEE RA hexadecimal
format with the octets separated by hyphens. E.g. `AC-DE-48-23-45-67`.

###### Query parameters
* `crc` mac address's CRC code, only valid with Gigaset provider.

###### Body
A JSON object with the `url` field:
* `url` URL of configuration server, must be formatted as URI standard https://tools.ietf.org/html/rfc3986.

Example:
```json
{"url":"https://example.com/"}
```
#### Response
The API returns a HTTP status code `200` in case of success.

For providers other than `ymcs`, the response body is empty.

For the `ymcs` provider, the response body is a JSON object:
* `device_pin`: a string PIN if returned by YMCS, otherwise `null`.

In case of error, the API returns a JSON object.
JSON object fields:
* `error`: specific error code.
* `message`: additional information (optional).
##### 200
The device was configured successfully

YMCS response example:
```json
{"device_pin":"123456"}
```

##### 400
Errors codes:
* `missing_mac_address`: the mac address of the device is missing.
* `malformed_mac_address`: the mac address of the device is malformed.
* `missing_url`: the url to associate to device is missing.
* `unable_to_parse_url`: the url is not correctly formatted.
* `unsupported_url_scheme`: the scheme of the url is not supported (valid schemes: `ftp`, `tftp`, `http`, `https`)
* `missing_mac-id_crc`: the crc code of the provided mac address is missing,
  error returned only in case of Gigaset provider.
* `invalid_mac-id_crc_format`: the crc code provided is in invalid format,
  error returned only in case of Gigaset provider.

##### 404
Errors codes:
* `provider_not_supported`: the selected providers is not supported or disabled.

##### 500
Errors codes:
* `connection_to_remote_provider_failed`: the connection to remote provider failed, the addition field
  `message` is provided.
* `provider_remote_call_failed`: the remote provider responded with a HTTP status code that is not `200`.
* `read_remote_response_failed`: error on read remote response.
* `unknown_response_from_provider`: unknown response from remote provider.
* `malformed_url`: the provided url was not accepted by the remote provider.
* `not_valid_mac_address`: the provided mac address was not accepted by the remote provider.
* `device_owned_by_other_user`: the provided mac address was already configured by another provider user.
* `unknown_provider_error`: the error returned by provider is unknown, the addition field
  `message` is provided.

## Usage

```
Usage of ./falconieri:
  -c string
    	Path to configuration file (default "/opt/falconieri/conf.json")
```

## Configuration
Falconieri can configured via json file or environment variables, if present **the values in the environment variable take the precedence** over the ones declared in files.

### JSON

#### Providers configuration
The `provider` section define the remote providers configuration, common configurations fields to each provider:

* `user` Username for access to provider
* `password` Password for access to provider 
* `rpc_url` The URL for XML-RPC requests
* `disable` Enable/Disable the provider, default `false`

Gigaset specific:

* `disable_crc` If set to `true` Falconieri don't send the mac address's CRC code, default `false`

YMCS, GRAPE and SRAPS specific:

* `base_url` API base URL, for example:
  * `https://eu-api.ymcs.yealink.com` for YMCS provider
  * `https://api.grape.gigaset.net/api/v1/` for GRAPE provider
  * `https://api.sraps.snom.com/api/v1/` for SRAPS provider
* `client_id` Client ID
* `client_secret` Client secret/key

Example:

```json
{
  "providers": {
     "snom": {
       "user":"user",
       "password": "password",
       "rpc_url": "https://secure-provisioning.snom.com:8083/xmlrpc/",
       "disable": false
     },
     "ymcs": {
       "base_url": "https://eu-api.ymcs.yealink.com",
       "client_id": "your-client-id",
       "client_secret": "your-client-secret",
       "disable": false
     },
     "grape": {
       "base_url": "https://api.grape.gigaset.net/api/v1/",
       "client_id": "your-client-id",
       "client_secret": "your-client-secret",
       "disable": false
     },
     "sraps": {
       "base_url": "https://api.sraps.snom.com/api/v1/",
       "client_id": "your-client-id",
       "client_secret": "your-client-secret",
       "disable": false
     }
  }
}
```

### Environment Variables

* `SNOM_USER` Username for access to Snom provider
* `SNOM_PASSWORD` Password for access to Snom provider
* `SNOM_RPC_URL` The URL for XML-RPC requests of Snom provider
* `SNOM_DISABLE` Enable/Disable the provider, default `false`

* `GIGASET_USER` Username for access to Gigaset provider
* `GIGASET_PASSWORD` Password for access to Gigaset provider
* `GIGASET_RPC_URL` The URL for XML-RPC requests of Gigaset provider
* `GIGASET_DISABLE_CRC` If set to `true` Falconieri don't send the mac address's CRC code, default `false`
* `GIGASET_DISABLE` Enable/Disable the provider, default `false`

* `FANVIL_USER` Username for access to Fanvil provider
* `FANVIL_PASSWORD` Password for access to Fanvil provider
* `FANVIL_RPC_URL` The URL for XML-RPC requests of Fanvil provider
* `FANVIL_DISABLE` Enable/Disable the provider, default `false`

* `YEALINK_USER` Username for access to Yealink legacy provider
* `YEALINK_PASSWORD` Password for access to Yealink legacy provider
* `YEALINK_RPC_URL` The URL for XML-RPC requests of Yealink legacy provider
* `YEALINK_DISABLE` Enable/Disable the provider, default `false`

* `YMCS_BASE_URL` YMCS API base URL, for example `https://eu-api.ymcs.yealink.com`
* `YMCS_CLIENT_ID` OAuth client ID issued by Yealink
* `YMCS_CLIENT_SECRET` OAuth client secret issued by Yealink
* `YMCS_DISABLE` Enable/Disable the YMCS provider, default `false`

* `GRAPE_BASE_URL` Grape API base URL, for example `https://api.grape.gigaset.net/api/v1/`
* `GRAPE_CLIENT_ID` Hawk id issued by Gigaset for HMAC authentication
* `GRAPE_CLIENT_SECRET` Hawk key issued by Gigaset for HMAC authentication
* `GRAPE_DISABLE` Enable/Disable the Grape provider, default `false`

* `SRAPS_BASE_URL` SRAPS API base URL, for example `https://api.sraps.snom.com/api/v1/`
* `SRAPS_CLIENT_ID` Hawk id issued by SNOM for HMAC authentication
* `SRAPS_CLIENT_SECRET` Hawk key issued by SNOM for HMAC authentication
* `SRAPS_DISABLE` Enable/Disable the SRAPS provider, default `false`

## Other projects

[Tancredi](https://nethesis.github.io/tancredi/) is a phone provisioning engine ideal for internet deployments.

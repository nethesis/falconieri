# Falconieri

Remote Provisioning Gateway

Supported providers:

* [**Snom**](https://service.snom.com/display/wiki/XML-RPC+API)
* [**Gigaset**](https://teamwork.gigaset.com/gigawiki/display/GPPPO/Gigaset+Redirect+server)
* **Fanvil**
* [**Yealink**](http://support.yealink.com/documentFront/forwardToDocumentDetailPage?documentId=257)

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
* `url` URL of configuration server.

Example:
```json
{"url":"https://example.com/"}
```
#### Response
The API return a HTTP status code `200` with an empty body in case of success
or a json object in case of error.
Json  object field:
* `error`: specific error code.
* `message`: additional informations (optional).
##### 200
The device was configured successfully

##### 400
Errors codes:
* `missing_mac_address`: the mac address of the device is missing.
* `malformed_mac_address`: the mac address of the device is malformed.
* `missing_url`: the url to associate to device is missing.
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


Example:

```json
{
  "providers": {
     "snom": {
       "user":"user",
       "password": "password",
       "rpc_url": "https://secure-provisioning.snom.com:8083/xmlrpc/",
       "disable": false
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

* `YEALINK_USER` Username for access to Yealink provider
* `YEALINK_PASSWORD` Password for access to Yealink provider
* `YEALINK_RPC_URL` The URL for XML-RPC requests of Yealink provider
* `YEALINK_DISABLE` Enable/Disable the provider, default `false`

# Falconieri

Remote Provisioning Gateway

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

Supported providers:

* [**Snom**](https://service.snom.com/display/wiki/XML-RPC+API)
* [**Gigaset**](https://teamwork.gigaset.com/gigawiki/display/GPPPO/Gigaset+Redirect+server)
	* `disable_crc` If set to `true` Falconieri don't send the mac address's CRC code, default `false`
* [**Fanvil**]
* [**Yealink**](http://support.yealink.com/documentFront/forwardToDocumentDetailPage?documentId=257)

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

* `SNOM_USER` Username for access to snom provider
* `SNOM_PASSWORD` Password for access to snom provider
* `SNOM_RPC_URL` The URL for XML-RPC requests of snom provider
* `SNOM_DISABLE` Enable/Disable the provider, default `false`

* `GIGASET_USER` Username for access to snom provider
* `GIGASET_PASSWORD` Password for access to snom provider
* `GIGASET_RPC_URL` The URL for XML-RPC requests of snom provider
* `GIGASET_DISABLE_CRC` If set to `true` Falconieri don't send the mac address's CRC code, default `false`
* `GIGASET_DISABLE` Enable/Disable the provider, default `false`

* `FANVIL_USER` Username for access to snom provider
* `FANVIL_PASSWORD` Password for access to snom provider
* `FANVIL_RPC_URL` The URL for XML-RPC requests of snom provider
* `FANVIL_DISABLE` Enable/Disable the provider, default `false`

* `YEALINK_USER` Username for access to snom provider
* `YEALINK_PASSWORD` Password for access to snom provider
* `YEALINK_RPC_URL` The URL for XML-RPC requests of snom provider
* `YEALINK_DISABLE` Enable/Disable the provider, default `false`

## APIs

### PUT /providers/:provider/:mac
---

Register device on remote provisioning service, if the device is already configured,
the new configuration override the old.

#### Headers
* `Content-Type: application/json`

#### Path variables
* `provider`: Name of the remote provider
* `mac`: Mac address of the device

#### Query prameters
* `crc` mac address's CRC code, only valid with Gigaset provider.

#### Body
A JSON object with the `url` key:
* `url` URL of configuration server

Example:
```json
{"url":"https://example.com/"}
```



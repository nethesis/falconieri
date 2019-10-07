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

Supported providers:

* [**snom**](https://service.snom.com/display/wiki/XML-RPC+API)

Example:

```json
{
  "providers": {
     "snom": {
       "user":"user",
       "password": "password",
       "rpc_url": "https://secure-provisioning.snom.com:8083/xmlrpc/"
     }
}
```

### Environment Variables

* `SNOM_USER` Username for access to snom provider
* `SNOM_PASSWORD` Password for access to snom provider
* `SNOM_RPC_URL` The URL for XML-RPC requests of snom provider

## APIs

### PUT /providers/:provider/:mac
---

Register device on remote provisioning service, if the device is already configured,
the new configuration override the old, supported providers:

#### Headers
* `Content-Type: application/json`

#### Path variables
* `provider`: Name of the remote provider
* `mac`: Mac address of the device

#### Body
A JSON object with the `url` key:
* `url` URL of configuration server

Example:
```json
{"url":"https://example.com/"}
```



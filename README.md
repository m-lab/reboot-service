[![GoDoc](https://godoc.org/github.com/m-lab/reboot-service?status.svg)](https://godoc.org/github.com/m-lab/reboot-service) [![Build Status](https://travis-ci.org/m-lab/reboot-service.svg?branch=master)](https://travis-ci.org/m-lab/reboot-service) [![Coverage Status](https://coveralls.io/repos/github/m-lab/reboot-service/badge.svg?branch=master)](https://coveralls.io/github/m-lab/reboot-service?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/m-lab/reboot-service)](https://goreportcard.com/report/github.com/m-lab/reboot-service)

# Reboot API
This API allows to execute common operations on M-Lab platform's BMC
modules and CoreOS hosts.

It retrieves credentials for BMCs from Google Cloud Datastore, while access to
CoreOS hosts is granted through a private SSH key.

## Rebooting nodes with the Reboot API
The API provides a single endpoint, `/v1/reboot`, which allows to reboot a node with two different methods.

### POST /v1/reboot
Parameter         | Description
------------------| ----------------
`host`            | hostname to reboot
`method`          | `host` or `bmc`. Defaults to `bmc`. <br>

**Examples**

*Reboot mlab1.lga0t via the BMC:*
```
curl -X POST https://<reboot-api-url>/v1/reboot?host=mlab1.lga0t
```

*Reboot mlab1.lga0t by running `systemctl reboot` on CoreOS:*
```
curl -X POST https://<reboot-api-url>/v1/reboot?host=mlab1.lga0t&method=host
```

## Running the Reboot API

### Authenticating to Google Cloud Datastore

To fetch credentials to authenticate to the BMCs, the Reboot API needs to have
access to Google Cloud Datastore.

To do so, the Reboot API will use the credentials configured in
[gcloud](https://cloud.google.com/sdk/gcloud/reference/auth/login)
(if available) or the [GOOGLE_APPLICATION_CREDENTIALS](https://cloud.google.com/docs/authentication/production) mechanism,
also known as *Application Default Credentials*.

### Command line flags

All the command line flags can also be provided via a corresponding environment variable.

```
  -auth.password string
        Password for HTTP basic auth
  -auth.username string
        Username for HTTP basic auth
  -datastore.namespace string
        GCD namespace (default "reboot-api")
  -datastore.project string
        GCD project ID (default "mlab-sandbox")
  -listenaddr string
        Address to listen on (default ":8080")
  -prometheusx.listen-address string
         (default ":9990")
  -reboot.bmcport int
        DRAC port to use (default 806)
  -reboot.key string
        SSH private key path
  -reboot.sshport int
        SSH port to use (default 22)
  -reboot.user string
        User for rebooting CoreOS hosts (default "reboot-api")
  -tls.certs-dir string
        Folder where to cache TLS certificates (default "/var/tls/")
  -tls.host string
        Enable TLS and get LetsEncrypt certificate for this hostname```
```

### Development
To run the Reboot API locally, for development/testing:
- Build the Reboot API with  `go test ./... && go build`
- Run it with `./reboot-service`

Please note that by default the Reboot API will not require any authentication,
thus this method is **not suitable for production use**.

To configure HTTP Basic Authentication, you need to specify `-auth.username` and
`-auth.password`.

To reboot nodes via CoreOS, a valid SSH private key must be provided,
for example: `./reboot-service --reboot.key=/path/to/private.key` .

### Running with Docker
The Docker image can be generated with `docker -t reboot-api .`

After generating the image, you can run it with `docker run reboot-api`

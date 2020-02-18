# Reboot API

[![GoDoc](https://godoc.org/github.com/m-lab/reboot-service?status.svg)](https://godoc.org/github.com/m-lab/reboot-service) [![Build Status](https://travis-ci.org/m-lab/reboot-service.svg?branch=master)](https://travis-ci.org/m-lab/reboot-service) [![Coverage Status](https://coveralls.io/repos/github/m-lab/reboot-service/badge.svg?branch=master)](https://coveralls.io/github/m-lab/reboot-service?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/m-lab/reboot-service)](https://goreportcard.com/report/github.com/m-lab/reboot-service)

This API allows to execute common operations on M-Lab platform's BMC
modules and CoreOS hosts.

It retrieves credentials for BMCs from Google Cloud Datastore, while access to
CoreOS hosts is granted through a private SSH key.

## Rebooting nodes with the Reboot API

The API provides a reboot endpoint, `/v1/reboot`, which allows to reboot a node with two different methods.

### POST /v1/reboot

Parameter         | Description
------------------| ----------------
`host`            | hostname to reboot
`method`          | `host` or `bmc`. Defaults to `bmc`.

#### Examples

*Reboot mlab1.lga0t via the BMC:*

```bash
curl -X POST https://<reboot-api-url>/v1/reboot?host=mlab1.lga0t
```

*Reboot mlab1.lga0t by running `systemctl reboot` on CoreOS:*

```bash
curl -X POST https://<reboot-api-url>/v1/reboot?host=mlab1.lga0t&method=host
```

## End-to-end testing 

The `/v1/e2e` endpoint allows to run an e2e test on a specific BMC.

This endpoint returns a valid Prometheus metric representing the status of the BMC:

```reboot_e2e_result{status="<status>",target="<hostname>"} 1```

Possible statuses are:

Status         | Description
- | -
ok | Connection to this BMC was successful
credentials_not_found | Credentials to access this BMC are not available in the Credentials store
connection_failed | Connection to this BMC failed

Results are cached by default. You can configure the cache capacity and TTL with `-e2e.cache-capacity` and `-e2e.cache-ttl`.

### GET /v1/e2e

Parameter         | Description
------------------| ----------------
`target`          | hostname of the BMC to check


#### Examples

```bash
curl https://<reboot-api-url>/v1/e2e?target=mlab1d.lga0t.measurement-lab.org
```

*Output*:
```
# HELP reboot_e2e_result E2E test result for this target
# TYPE reboot_e2e_result gauge
reboot_e2e_result{status="ok",target="mlab1d.lga0t.measurement-lab.org"} 1
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

For a list of the available flags, run

```bash
./reboot-service -h
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

- Build the docker image
  - `docker -t reboot-api .`

- Run it
  - `docker run reboot-api`

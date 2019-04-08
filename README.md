# reboot-service
This reboot service allows to execute common operations on M-Lab platform's
BMC module (currently DRAC).

It retrieves login credentials from Google Cloud Datastore.

### Things that work:
- Get DRAC credentials from Datastore
- Authentication to Datastore with custom credentials
(through the `GOOGLE_APPLICATION_CREDENTIALS` environment variable)
- Log into DRACs with username/password
- Execute commands on DRACs
- Execute a reboot command

### Things that do *NOT* work yet:
- Send commands with more than 256 characters
- SSH host key validation (currently skipped - **UNSAFE**)

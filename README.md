# reboot-service
This reboot service allows to execute common operations on M-Lab platform's
DRACs.

It retrieves login credentials from Google Cloud Datastore, which replaces PLC
as credential store.

### Things that work:
- Get DRAC credentials from Datastore
- Log into DRACs with username/password
- Execute commands on DRACs
- Execute a reboot command
    
### Things that do *NOT* work yet:
- Authentication to Datastore with custom application credentials
- Send commands with more than 256 characters
- SSH host key validation (currently skipped - **UNSAFE**)

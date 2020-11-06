# Group Member Dump

Dump members of LDAP groups to a CSV file.

## How to build

Run the command:
```sh
make build
```

The binary for current OS and ARCH can be found in `bin` directory.

## Usage

```sh
./gmdump -h
Usage of ./gmdump:
  -H, --host string       LDAP server to query against (default "localhost")
  -u, --username string   The full username with domain to bind with (e.g. 'user@example.com')
  -p, --password string   Password to use. If not specified, will be prompted for
      --secure            Use LDAPS. This will not verify TLS certs, however. (default: false)
  -b, --basedn string     DN of organizational unit or group to dump members from
  -o, --output string     Save results to file
      --attrs strings     Comma separated attributes to dump (default [cn,mail])
  -v, --version           Show version and exit
```

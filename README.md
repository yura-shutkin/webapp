# Web app

This is a simple application that shows environment variables and values from files in specified directories

Initially written for workshop about hashicorp vault integration into k8s cluster

## ENV_VARS

| variable name     | expected value                             | comment                                                                        |
|-------------------|--------------------------------------------|--------------------------------------------------------------------------------|
| LISTEN_ADDR       | `0.0.0.0:8080`                             | On which IP addr and port application should be launched                       |
| SECRETS_DIRS      | `/vault/secrets,/opt/secrets`              | List of dirs on which look for files/secrets                                   |
| HTTP_HOSTS        | `http://google.com;http://api.ns.svc:8080` | list of hosts which application should try to reach for testing network access |
| HTTP_CHECK_PERIOD | `5`                                        | How often application should try to reach provided HTTP Hosts                  |

## Locations

| Location     | Usage                                                                           |
|--------------|---------------------------------------------------------------------------------|
| `/`          | Main Web UI of the application. Shows env vars and values from files            |
| `/json`      | Return env vars and values of files founded in dirs from `SECRETS_DIRS` env var |
| `/ping`      | Response 200 OK                                                                 |
| `/net-check` | Perform network access check to hosts provided in `HTTP_HOSTS` env var if any   |
| `/metrics`   | Prometheus metrics                                                              |

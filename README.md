# go_proxy

This project implements a proxy server using Go and the `goproxy` package. The server redirects requests to a specified target server and logs requests and responses.

## Features

- Redirects requests to a specified target server
- Logs requests and responses to separate files
- Configurable via a YAML file

## Configuration

Create a `config.yaml` file in the project root directory with the following content:

```yaml
port: "8080"
target_server: "<target url>"
```

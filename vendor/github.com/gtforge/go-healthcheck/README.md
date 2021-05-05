# Go Healthcheck

[![codecov](https://codecov.io/gh/gtforge/go-healthcheck/branch/master/graph/badge.svg?token=Qri4MTocQi)](https://codecov.io/gh/gtforge/go-healthcheck)
[![Build Status](https://travis-ci.com/gtforge/go-healthcheck.svg?token=4DXoanrMBg3g2xznPprL&branch=master)](https://travis-ci.com/gtforge/go-healthcheck)

Repository `go-healthcheck` provides a service and an http handler for healthcheck your application. 
In common Gett infrastructure should be mount on `/alive` endpoint.

## Components

Healthcheck consists of Service, which is responsible for calling Pingers — functions to report stuff
and provide errors.

### Health Check Service

Implementation of Alive service which is a requirement for all Gett services and prerequisite for k8s. Is build with default reporter.

To instantiate a service one should provide variable list of pingers.

```go
healthcheck.NewHealthCheck(
    healthcheck.MakeDbPinger(db, "main_db"), // The built in handler to check DB connection 
    func(ctx context.Context) (map[string]interface{}, error) {
        // your check code and reporting, returning error means that service is not alive
        return map[string]interface{}{"gop": "stop"}, nil
    },
)
```

#### Build

In order to inject **build time** and **commit** you should use build time linking.

##### You use go modules

Just provide import path as you see it. 

```
go build -ldflags "-X github.com/gtforge/go-healthcheck.Buildstamp=`date -u +%Y/%m/%d_%H:%M:%S` -X github.com/gtforge/go-healthcheck.Commit=`git rev-parse HEAD`"
```

##### You use vendoring or dep
May be a little more tricky import path should be with `vendor` in place. Like this

```
go build -ldflags "-X github.com/you_name/repo_name/vendor/github.com/gtforge/go-healthcheck.Buildstamp=`date -u +%Y/%m/%d_%H:%M:%S` -X github.com/you_name/repo_name/vendor/github.com/gtforge/go-healthcheck.Commit=`git rev-parse HEAD`"
```
The fastest possible way to know the exact import path is
`go tool nm backend | grep go-healthcheck` where `backend` is freshly build binary

### Pingers (reporters)

Pinger is a function, taking `context.Context` and returning `map[string]interface{}` and `error`.
Note, that if pinger returned error, map will still be merged to output.

```go
func (ctx context.Context) (resp map[string]interface{},err error) {
	resp["cats_count"] = 42
	return
}
```

#### Default pinger

Currently provided default pinger, for common information:

```
{
    "commit":        Commit,
    "build_time":    Buildstamp,
    "startup_time":  time.Now(),
    "num_cpu":       runtime.NumCPU(),
    "num_goroutine": runtime.NumGoroutine(),
    "go_version":    runtime.Version(),
}
```

#### DB Pinger

And helper DB pinger, to check database, based on method `PingContext`

`healthcheck.MakeDbPinger(db, "main_db")`

First param — db connection, second - name for reporting.

_Please note, that in current version there is no response keys collision prevention,
so it's developer's responsibility to check uniqueness of response keys._

#### Git Branch pinger

Git branch pinger works similar to default one

Build:
```bash
go build -ldflags "-X github.com/gtforge/go-healthcheck.GitBranch=`git rev-parse --abbrev-ref HEAD`" 
```
Add to healthcheck
```go
hc := healthcheck.NewHealthCheck(healthcheck.MakeBranchPinger)
```

Also method which will fail if branch was not provided
```go
pinger, err := MakeBranchPingerWithError
if err != nil {
	log.Fatal("Git branch pinger failed")
}
```

### Handler

Handler is made out of Healthcheck service and can be mount later to your router.

`handler := healthcheck.MakeHealthcheckHandler(hc)`

#### Usage

Instantiate a Health Check service. You can provide built in or custom checkers-reporters.

```go
package main

import (
    "net/http"
    "context"
    
    "github.com/gtforge/go-healthcheck"
)

func main() {
    // You can provide a variable list of checkers.
    hc := healthcheck.NewHealthCheck(
        healthcheck.MakeDbPinger(db, "master"), // The built in handler to check DB connection
        func(ctx context.Context) (map[string]interface{}, error) {
		    // your check code and reporting, returning error means that service is not alive
            return map[string]interface{}{"processing": "stopped"}, nil
        },
    )

    handler := healthcheck.MakeHealthcheckHandler(hc)
    mux := http.NewServeMux()
    mux.Handle("/alive", handler)
    http.ListenAndServe(":3000", mux)
}
```

Output will be something like:
```json
{
  "alive": true,
  "build_time": "",
  "startup_time": "2019-10-10T12:28:00.11228+03:00",
  "commit": "",
  "go_version": "go1.11.4",
  "processing": "stopped",
  "num_cpu": 8,
  "num_goroutine": 3
}
```

## Owner

Vladislav Bogomolov [bogomolov@gett.com](mailto:bogomolov@gett.com)


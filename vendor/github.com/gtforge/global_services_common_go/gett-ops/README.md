## General

A collection of monitoring, status and error reporting tools for Gett services.

What's inside:
* Access Log enable
* Beego Settings
* Airbrake configured for async error reporting

### Access Log
for enabling access log for your service
```go
import "github.com/gtforge/services_common_go/gett-ops/gett-beego"

gettBeego.AccessLog = true

```

### Beego Settings
It is possible to override the default server listen timeout (if not specified default is 15 seconds).
Setting it to 0 means no timeout (Use with extreme caution!)
In `base.yml` :

```yaml
  global:
  # ...
  api:
    # ...
    server_listen_timeout: 15
    # ...
```

### Airbrake
Provide an API key using ENV params AIRBRAKE_PROJECT_ID and AIRBRAKE_API_KEY
OR
Provide an API key using `config/airbrake.yml`:

```yaml
  # ...
  apikey: APIKEY
  project_id: PROJECTID
  # ...
```


### New Relic Instrumentation
if you have manual code blocks that you want to see in new relic as transactions/ datastore operations, you can use instrumentation API.

#### Examples
```go
    gettOps.InstrumentBlock("myBlockName", func() {
    	//do something hard
    })
```

```go
    gettOps.InstrumentDatastore("BulkUpdate", "RoutingSettings", func() error {
    	//Bulk DB operation...
    	return nil
    })
```

### New Relic Distributed Tracing
In order to enable distributed tracing for your service, make sure you set the following environment variable:

`NEW_RELIC_DISTRIBUTED_TRACING_ENABLED = true`

#### Documentation
https://docs.newrelic.com/docs/apm/distributed-tracing/ui-data/understand-use-distributed-tracing-data

### Errors handling & logging
`gett-ops` supports standard beego errors handling and uses logrus for sending there errors to airbrake / kibana / newrelic.

The default log level is error (`logrus.ErrorLevel`), but you can change it in this way:
```
func (c *BaseController) AbortWithWarn(code string) {
	c.Ctx.Input.SetData(gettBeego.ErrorLogLevel, logrus.WarnLevel)
 	c.Abort(code)
}
```

### Pprof
Each service supports a graphic visualization for profiling purposes
#### Preconditions
1. Install pprof
```bash
go get github.com/google/pprof
```
2. Install GraphViz. For mac users you can use homebrew
```bash
brew install Graphviz
```

#### Usage
* Replace `<service_url>` with you service address (pay attention, gibberish below is not a mitake).
* Replace `<type>` with the type of profiling you want. Options include `heap`, `mutex`, `goroutine`, `block`

```bash
pprof -http=localhost:8080 http://gtdeploy:gtdeploy%21%40%2A@<service_url>/debug/pprof/<type> && open http://localhost:8080
```
For example:
```bash
pprof -http=localhost:8080 http://gtdeploy:gtdeploy%21%40%2A@employees-scrum75.gett.io/debug/pprof/heap && open http://localhost:8080
```


package healthcheck

import (
	"context"
	"runtime"
	"time"

	"strings"

	"github.com/pkg/errors"
)

var (
	// go build -i -v -ldflags "-X github.com/gtforge/go-healthcheck.Buildstamp=`date -u +%Y/%m/%d_%H:%M:%S` -X github.com/gtforge/go-healthcheck.Commit=`git rev-parse HEAD`"
	Buildstamp string
	Commit     string
)

// Type of checker-reporter for HealthCheck, returns map of messages to report and error
// If error isn't nil the service is considered dead.
type Pinger func(ctx context.Context) (map[string]interface{}, error)

type Checker interface {
	Check(ctx context.Context) (map[string]interface{}, error)
}

// HealthCheck service
type HealthCheck struct {
	Pingers []Pinger
}

// Exact logic
func (hc HealthCheck) Check(ctx context.Context) (map[string]interface{}, error) {
	response := map[string]interface{}{}
	var statusError error
	for _, pinger := range hc.Pingers {
		resp, err := pinger(ctx)
		if err != nil {
			if statusError == nil {
				statusError = err
			} else {
				statusError = errors.Wrap(statusError, err.Error())
			}
		}
		for k, v := range resp {
			response[k] = v
		}
	}
	response["alive"] = true
	if statusError != nil {
		response["alive"] = false
		response["errors"] = strings.Replace(statusError.Error(), ": ", "; ", len(hc.Pingers))
	}
	return response, statusError
}

// Factory of HealthCheck service
func NewHealthCheck(pingers ...Pinger) *HealthCheck {
	hc := HealthCheck{
		Pingers: append(pingers, DefaultPinger(time.Now())),
	}
	return &hc
}

// Default reporter
func DefaultPinger(startup time.Time) Pinger {
	return func(_ context.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			"commit":        Commit,
			"build_time":    Buildstamp,
			"startup_time":  startup,
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
			"go_version":    runtime.Version(),
		}, nil
	}
}

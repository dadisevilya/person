package gettOps

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gtforge/global_services_common_go/gett-config"

	"github.com/pkg/errors"
	"gopkg.in/airbrake/gobrake.v2"
)

type AirbrakeAgent struct {
	notifier *gobrake.Notifier
}

// for info check github.com/pkg/errors/stack.go:80
type stackTracer interface {
	StackTrace() errors.StackTrace
}

func NewAirbrakeAgent(airbrakeProjectID int64, airbrakeKey string) *AirbrakeAgent {
	a := &AirbrakeAgent{
		notifier: gobrake.NewNotifier(airbrakeProjectID, airbrakeKey),
	}

	a.notifier.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
		if os.Getenv("PARAM_ENV") == "" {
			notice.Context["environment"] = gettConfig.Settings.AppEnv
			return notice
		}
		notice.Context["environment"] = os.Getenv("PARAM_ENV")
		return notice
	})

	return a
}

func (a *AirbrakeAgent) Notify(e interface{}, req *http.Request) {
	if a == nil {
		return
	}
	notice := a.NewNotice(e, req, 3)
	a.notifier.SendNoticeAsync(notice)
}

func (a *AirbrakeAgent) NotifyNotice(notice *gobrake.Notice) {
	if a == nil {
		return
	}
	a.notifier.SendNoticeAsync(notice)
}

func (a *AirbrakeAgent) NewNotice(e interface{}, req *http.Request, depth int) *gobrake.Notice {
	notice := gobrake.NewNotice(e, req, depth)
	if e, ok := e.(stackTracer); ok {
		notice.Errors[0].Backtrace = make([]gobrake.StackFrame, len(e.StackTrace()))
		for i, f := range e.StackTrace() {
			// for info check github.com/pkg/errors/stack.go:40
			li, err := strconv.Atoi(fmt.Sprintf("%d", f))
			if err != nil {
				li = -1 // this won't happen, but just in case
			}
			notice.Errors[0].Backtrace[i] = gobrake.StackFrame{
				// nolint
				File: fmt.Sprintf("%s", f),
				Line: li,
				// nolint
				Func: fmt.Sprintf("%n", f),
			}
		}
	}

	return notice
}

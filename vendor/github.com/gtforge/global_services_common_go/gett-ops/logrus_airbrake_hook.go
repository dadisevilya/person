package gettOps

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/newrelic/go-agent"
	"github.com/pkg/errors"
	"github.com/gtforge/gls"
	"gopkg.in/airbrake/gobrake.v2"
)

func init() {
	logrus.AddHook(NewLogrusAirbrakeHook())
}

type logrusAirbrakeHook struct {
}

func NewLogrusAirbrakeHook() logrus.Hook {
	return &logrusAirbrakeHook{}
}

func (l *logrusAirbrakeHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel} //Do we need Warnings here?
}

func (l *logrusAirbrakeHook) Fire(entry *logrus.Entry) error {
	var notice *gobrake.Notice
	var req *http.Request

	//this one is from the filter
	if glsReq := gls.Get("http_request"); glsReq != nil {
		if v, ok := glsReq.(*http.Request); ok {
			req = v
		}
	}

	//this one is if you want to override it
	if val, ok := entry.Data["req"]; ok {
		if r, ok := val.(*http.Request); ok {
			req = r
			delete(entry.Data, "req")
		}
	}

	// https://github.com/gemnasium/logrus-airbrake-hook/blob/master/airbrake.go
	var notifyErr error
	err, ok := entry.Data["error"].(error)
	if ok {
		notifyErr = err
	} else {
		notifyErr = errors.New(entry.Message)
	}

	//this one reports to NR
	if tx := gls.Get(GlsNewRelicTxnKey); tx != nil {
		if nrtx, ok := tx.(newrelic.Transaction); ok {
			_ = nrtx.NoticeError(notifyErr)
		}
	}

	notice = Airbrake.NewNotice(notifyErr, req, 3)

	// adding fields to notice
	for k, v := range entry.Data {
		notice.Context[k] = fmt.Sprint(v)
	}

	notice.Context["level"] = entry.Level.String()
	Airbrake.NotifyNotice(notice)
	return nil
}

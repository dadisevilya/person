package gettOps

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"

	"github.com/newrelic/go-agent"
	"github.com/pkg/errors"
	"github.com/gtforge/gls"
)

type middlewares struct {
}

func (m middlewares) setContext() func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	return func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		gls.Set("request", req)

		next(res, req)

		gls.Cleanup()
	}
}

func (m middlewares) recovery() func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	return func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, true)]
				Airbrake.Notify(err, req)
				log.Printf("PANIC: %s\n", err)
				log.Println(string(stack))
				if txn := gls.Get(GlsNewRelicTxnKey); txn != nil {
					tx := txn.(newrelic.Transaction)
					_ = tx.NoticeError(errors.New(err.(string)))
				}
				gls.Cleanup()
				body := []byte("500 Internal Server Error")
				res.WriteHeader(http.StatusInternalServerError)
				_, _ = res.Write(body)
			}
		}()
		next(res, req)
	}
}

func (m middlewares) cors() func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	return func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		if os.Getenv("APP_ENV") == "" || os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "stage" {
			res.Header().Set("Access-Control-Allow-Origin", "*")
			res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
			res.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		next(res, req)
	}
}

func (m middlewares) newRelicRecorder() func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	return func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		re := regexp.MustCompile("[0-9]{2,}")
		path := re.ReplaceAllString(req.URL.Path, ":id")
		txName := fmt.Sprintf("%s %s", req.Method, path)
		tx := Newrelic.StartTransaction(txName, res, req)
		defer tx.End()
		gls.Set(GlsNewRelicTxnKey, tx)

		next(res, req)
	}
}

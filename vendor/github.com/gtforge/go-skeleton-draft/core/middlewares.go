package skeleton

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/gtforge/global_services_common_go/gett-ops"
	"github.com/gtforge/global_services_common_go/gett-utils/transactionid"
	"github.com/gtforge/gls"
	"github.com/newrelic/go-agent"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	userAgentHeaderName = "User-Agent"
)

// getAllMiddleware applies all middleware and returns one http.Handler to be registered on router
func getAllMiddleware(router http.Handler, logger *logrus.Logger) http.Handler {
	return newCorsMiddleware(
		setContextMiddleware(
			tracingMiddleware( // generate request_id and context
				newRecoveryMiddleware(logger)( // recover from panic()
					loggingMiddleware(logger)(
						newRelicRecorder(
							router,
						),
					),
				),
			),
		),
	)
}

func setContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		gls.Set("http_request", req)

		next.ServeHTTP(res, req)

		gls.Cleanup()
	})
}

func newRecoveryMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 1024*8)
					stack = stack[:runtime.Stack(stack, true)]
					gettOps.Airbrake.Notify(err, req)
					if _, isError := err.(error); isError {
						err = errors.WithStack(err.(error))
					}
					logger.Printf("PANIC: %s\n", err)
					logger.Println(string(stack))
					if txn := req.Context().Value(gettOps.GlsNewRelicTxnKey); txn != nil {
						tx := txn.(newrelic.Transaction)
						_ = tx.NoticeError(errors.New(err.(string)))
					}
					gls.Cleanup()
					body := []byte(http.StatusText(http.StatusInternalServerError))
					res.WriteHeader(http.StatusInternalServerError)
					_, _ = res.Write(body)
				}
			}()
			next.ServeHTTP(res, req)
		})
	}
}

func newCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if os.Getenv("APP_ENV") == "" || os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "stage" {
			res.Header().Set("Access-Control-Allow-Origin", "*")
			res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
			res.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		next.ServeHTTP(res, req)
	})
}

func newRelicRecorder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		re := regexp.MustCompile("[0-9]{2,}")
		path := re.ReplaceAllString(req.URL.Path, ":id")
		txName := fmt.Sprintf("%s %s", req.Method, path)
		tx := gettOps.Newrelic.StartTransaction(txName, res, req)
		gls.Set(gettOps.GlsNewRelicTxnKey, tx)
		ctx := context.WithValue(req.Context(), gettOps.GlsNewRelicTxnKey, tx)
		reqWithContext := req.WithContext(ctx)

		next.ServeHTTP(res, reqWithContext)

		_ = tx.End()
	})
}

func tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get(transactionid.RequestIdKey)
		if requestID == "" {
			requestID = uuid.NewV4().String()
		}
		ctx := context.WithValue(req.Context(), transactionid.RequestIdKey, requestID)
		gls.Set(transactionid.RequestIdKey, requestID)
		w.Header().Set(transactionid.RequestIdKey, requestID)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func loggingMiddleware(logger *logrus.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := statusWriter{ResponseWriter: w}
			next.ServeHTTP(&sw, r)

			if r.RequestURI == "/alive" {
				return
			}

			duration := time.Since(start)
			reqLog := LogEntry{
				Host:       r.Host,
				RemoteAddr: r.RemoteAddr,
				Method:     r.Method,
				RequestURI: r.RequestURI,
				Proto:      r.Proto,
				UserAgent:  r.Header.Get(userAgentHeaderName),
				Duration:   duration.String(),
				Status:     sw.status,
				ContentLen: sw.length,
				RequestID:  r.Context().Value(transactionid.RequestIdKey).(string),
			}

			entry, _ := json.Marshal(reqLog)
			logger.Print(string(entry))
		})
	}
}

const basicAuthUser = "gtdeploy"
const basicAuthPass = "gtdeploy!@*"

func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()

		if user != basicAuthUser || pass != basicAuthPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

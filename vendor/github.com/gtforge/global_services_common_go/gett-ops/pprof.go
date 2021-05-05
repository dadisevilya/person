package gettOps

import (
	"net/http"
	"net/http/pprof"
)

const PProfProtectedPath = "/debug/pprof/*"

func PProfHTTPHandlers() map[string]http.Handler {
	handlerMap := map[string]http.Handler{}

	handlerMap[PProfProtectedPath] = http.HandlerFunc(pprof.Index)
	handlerMap["/debug/pprof/cmdline"] = http.HandlerFunc(pprof.Cmdline)
	handlerMap["/debug/pprof/profile"] = http.HandlerFunc(pprof.Profile)
	handlerMap["/debug/pprof/symbol"] = http.HandlerFunc(pprof.Symbol)
	handlerMap["/debug/pprof/trace"] = http.HandlerFunc(pprof.Trace)

	return handlerMap
}

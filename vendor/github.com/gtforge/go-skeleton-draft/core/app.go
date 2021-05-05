package skeleton

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gtforge/global_services_common_go/gett-config"
	"github.com/gtforge/global_services_common_go/gett-ops"
	"github.com/gtforge/go-healthcheck"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type App struct {
	config  gettConfig.AppConfig
	ctx     context.Context
	handler http.Handler
	logger  *logrus.Logger
	pingers []healthcheck.Pinger
}

type BlankFormatter struct{}

// NewApp - creates new app instance with passed http handler and logger
// It also initializes the application config
func NewApp(config gettConfig.AppConfig, handler http.Handler, logger *logrus.Logger, pingers []healthcheck.Pinger) App {
	app := App{
		config:  config,
		handler: handler,
		logger:  logger,
		pingers: pingers,
	}

	gettOps.InitOps()

	log.SetFlags(0)
	log.Print(skeletonBanner)
	log.SetFlags(log.LstdFlags)

	return app
}

// Run starts the application
func (app *App) Run(httpTermination chan<- struct{}) {
	var cancel context.CancelFunc
	app.ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	httpServerErr := make(chan error)

	server, err := app.createHTTPServer()
	if err != nil {
		app.logger.Fatal("could not create http server, %v", err.Error())
	}

	go func() {
		app.logger.Printf("server is listening on %s", server.Addr)
		httpServerErr <- server.ListenAndServe()
	}()

	sigint := make(chan os.Signal, 1)
	defer close(sigint)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	select {
	case <-app.ctx.Done():
		app.logger.Printf("context cancellation %v \n", app.ctx.Err().Error())
	case err := <-httpServerErr:
		app.logger.Printf("could not run server: %s \n", err.Error())
	case sig := <-sigint:
		app.logger.Printf("signal received: %v \n", sig.String())
	}

	app.logger.Println("HTTP server is gracefully shutting down, waiting for active connections to finish")
	if err := server.Shutdown(app.ctx); err != nil {
		// Error from closing listeners, or context timeout:
		app.logger.Fatalf("could not gracefully shutdown the server: %s\n", err)
	}

	log.SetFlags(0)
	log.Print(maydayBanner)
	log.SetFlags(log.LstdFlags)

	httpTermination <- struct{}{}
}

func createHealthCheckHandler(pingers ...healthcheck.Pinger) http.Handler {
	hc := healthcheck.NewHealthCheck(pingers...)
	return healthcheck.MakeHealthcheckHandler(hc)
}

func (app *App) createHTTPServer() (*http.Server, error) {
	if app.handler == nil {
		return nil, errors.New("HTTP handler is not defined")
	}

	http.Handle("/alive", createHealthCheckHandler(app.pingers...))
	http.Handle("/debug/pprof", BasicAuthMiddleware(http.DefaultServeMux))

	port := os.Getenv("HTTP_PORT")
	if len(port) == 0 {
		if app.config.Env.IsDev() {
			port = "8080"
		} else {
			port = "80"
		}
	}

	return &http.Server{
		Handler:      getAllMiddleware(app.handler, app.logger),
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, nil
}

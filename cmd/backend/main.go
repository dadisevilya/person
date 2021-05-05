package main

import (
	"os"

	"github.com/gtforge/global_services_common_go/gett-config"
	gettOps "github.com/gtforge/global_services_common_go/gett-ops"
	"github.com/gtforge/global_services_common_go/gett-storages"
	skeleton "github.com/gtforge/go-skeleton-draft/core"
	"github.com/gtforge/gorm"
	"github.com/sirupsen/logrus"
)

type Deps struct {
	DB    *gorm.DB
	Redis *gettStorages.GtRedisClient
}

// initServices - initialize required infra packages (global_services_common_go)
func initServices(config gettConfig.AppConfig) Deps {
	gettStorages.InitDb(config.Db, config.AppEnv)
	gettStorages.InitRedis(config.Redis, config.AppEnv)

	return Deps{
		DB:    gettStorages.DB,
		Redis: gettStorages.RedisClient,
	}
}

func main() {
	// Initializing router and logger to be used by App instance
	logger := createLogger()
	config := gettConfig.GetConfig()

	// Initializing required infra dependencies
	deps := initServices(config)
	pingers := healthCheckPingers(deps.DB.DB())
	router := createRouter()
	app := skeleton.NewApp(config, router, logger, pingers)

	httpTermination := make(chan struct{})
	go app.Run(httpTermination)
	<-httpTermination
}

func createLogger() *logrus.Logger {
	logger := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.JSONFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        getLogLevel(),
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	logger.AddHook(gettOps.NewLogrusAirbrakeHook())

	return logger
}

func getLogLevel() logrus.Level {
	env := gettConfig.GetConfig().Env

	if env.IsProd() {
		return logrus.WarnLevel
	} else if env.IsStage() {
		return logrus.DebugLevel
	}

	return logrus.TraceLevel
}

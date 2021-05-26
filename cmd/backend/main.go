package main

import (
	"fmt"
	"github.com/gtforge/global_services_common_go/gett-mq"
	"github.com/gtforge/global_services_common_go/gett-mq/consumer"
	"github.com/gtforge/global_services_common_go/gett-mq/publisher"
	"github.com/gtforge/global_services_common_go/gett-workers"
	"github.com/gtforge/go-skeleton-draft/structure/pkg/events"
	"github.com/gtforge/go-skeleton-draft/structure/pkg/workers"
	"os"

	"github.com/gtforge/global_services_common_go/gett-config"
	"github.com/gtforge/global_services_common_go/gett-ops"
	"github.com/gtforge/global_services_common_go/gett-storages"
	"github.com/gtforge/go-skeleton-draft/core"
	"github.com/gtforge/gorm"
	"github.com/sirupsen/logrus"
)

type Deps struct {
	DB       *gorm.DB
	Redis    *gettStorages.GtRedisClient
	RabbitMQ *gettMQ.AMQPConnection
}

// initServices - initialize required infra packages (global_services_common_go)
func initServices(config gettConfig.AppConfig) Deps {
	gettStorages.InitDb(config.Db, config.AppEnv)
	gettStorages.InitRedis(config.Redis, config.AppEnv)
	consumer.InitMqConumer()
	return Deps{
		DB:       gettStorages.DB,
		Redis:    gettStorages.RedisClient,
		RabbitMQ: publisher.InitMqPublisher(),
	}
}

func main() {
	// Initializing router and logger to be used by App instance
	logger := createLogger()
	config := gettConfig.GetConfig()

	// Initializing required infra dependencies
	deps := initServices(config)
	releaseAllJobs()
	pingers := healthCheckPingers(deps.DB.DB())
	router := createRouter()
	app := skeleton.NewApp(config, router, logger, pingers)
	gettWorkers.InitJobsManager([]gettWorkers.Worker{
		workers.GetWorker(),
	}, map[string]string{"poll_interval": "1"})
	events.InitConsumer(deps.DB)

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

func releaseAllJobs() {
	logger := logrus.WithField("worker_name", "InMemoryWorker")
	keys, err := gettStorages.RedisClient.Keys(fmt.Sprintf("%v", "InMemoryWorker")).Result()
	if err != nil {
		logger.WithError(err).Error("[ReleaseAllJobs] error get worker keys")
		return
	}
	if len(keys) == 0 {
		logger.
			Info("[ReleaseAllJobs] nothing to release")
		return
	}
	count, err := gettStorages.RedisClient.Del(keys...).Result()
	if err != nil {
		logger.WithError(err).Error("[ReleaseAllJobs] error delete jobs keys")
		return
	}
	logger.
		WithField("count", count).
		Info("[ReleaseAllJobs] all jobs released")
}
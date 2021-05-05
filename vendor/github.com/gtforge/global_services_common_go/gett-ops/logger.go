package gettOps

import (
	"fmt"
	"io"
	"os"

	"github.com/gtforge/global_services_common_go/gett-config"

	log "github.com/sirupsen/logrus"
)

const defaultProdLogLevel = log.InfoLevel
const defaultDevLogLevel = log.DebugLevel
const logDir = "/app/src/log/"
const logFileNameExt = ".log"

func initLogger() {
	log.SetLevel(getLogLevel())
	log.SetOutput(getOutput())
	log.SetFormatter(&GettLogFormatter{})
}

func getLogDir() string {
	logDir := logDir
	if gettConfig.Settings.AppEnv == "development" || gettConfig.Settings.AppEnv == "test" {
		logDir = "./log/"
	}
	return logDir
}

func getOutput() io.Writer {

	if gettConfig.Settings.Env.Isk8s() {
		return io.Writer(os.Stdout)
	}

	logDir := getLogDir()
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic("Unexpected error - could not create logging directory: " + logDir)
	}

	logFilePath := extractLogFilePath(logDir)
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic("Unexpected error - could not open log file: " + logFilePath)
	}

	if gettConfig.Settings.AppEnv == "prod" {
		return logFile
	}

	return io.MultiWriter(os.Stdout, logFile)
}

func extractLogFilePath(logDir string) string {
	env := gettConfig.Settings.AppEnv
	if env == "prod" {
		env = "production" //Thanks Oleg...
	}

	return logDir + env + logFileNameExt
}

func getLogLevel() log.Level {
	level := defaultProdLogLevel
	if gettConfig.Settings.AppEnv != "prod" {
		level = defaultDevLogLevel
	}

	logLevel := gettConfig.Settings.GlobalSettings.GetString("logrus.log_level")
	if logLevel != "" {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			allLevels := ""
			for _, level := range log.AllLevels {
				allLevels += level.String() + ","
			}
			panic(fmt.Sprintf("Log level is invalid: %s, ca be only one of these: %s", logLevel, allLevels))
		}
		fmt.Println("Log level = " + logLevel)
		return level
	}

	return level
}

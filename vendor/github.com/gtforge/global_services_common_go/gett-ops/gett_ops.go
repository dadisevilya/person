package gettOps

import (
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/newrelic/go-agent"
	"github.com/spf13/viper"
	"fmt"
)

var Airbrake *AirbrakeAgent
var Newrelic newrelic.Application

const GlsNewRelicTxnKey = "newrelic_txn"

func InitOps() {
	initLogger()
	initAirbrake()
	initNewRelic()
}

func GetAllMiddlewares() []func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	m := []func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc){}
	m = append(m, middlewares{}.setContext())
	m = append(m, middlewares{}.cors())
	m = append(m, middlewares{}.newRelicRecorder())
	m = append(m, middlewares{}.recovery())
	return m
}

func initAirbrake() {
	var airbrakeProjectID int64
	var airbrakeKey string
	airbrakeProjectID, _ = strconv.ParseInt(os.Getenv("AIRBRAKE_PROJECT_ID"), 10, 64)
	airbrakeKey = os.Getenv("AIRBRAKE_API_KEY")
	if airbrakeKey == "" || airbrakeProjectID == 0 {
		settings := viper.New()
		settings.SetConfigFile(fmt.Sprintf("%s/airbrake.yml", getConfigFilePath()))
		err := settings.ReadInConfig()
		if err != nil {
			log.Println("Gett-ops airbrake error:", err)
		}
		airbrakeProjectID = int64(settings.GetFloat64("project_id"))
		airbrakeKey = settings.GetString("apikey")
	}

	if airbrakeKey != "" || airbrakeProjectID != 0 {
		Airbrake = NewAirbrakeAgent(airbrakeProjectID, airbrakeKey)
	}
}

func initNewRelic() {
	newRelicLicenseKey := os.Getenv("NEW_RELIC_LICENSE_KEY")
	newRelicApplicationName := os.Getenv("NEW_RELIC_APP_NAME")
	newRelicDistributedTracingEnabled := os.Getenv("NEW_RELIC_DISTRIBUTED_TRACING_ENABLED")

	if newRelicLicenseKey == "" || newRelicApplicationName == "" {
		settings := viper.New()
		settings.SetConfigFile(fmt.Sprintf("%s/secrets.yml", getConfigFilePath()))
		err := settings.ReadInConfig()
		if err != nil {
			log.Println("Gett-ops newrelic error:", err)
		}
		newRelicApplicationName = settings.GetString("newrelic_application_name")
		newRelicLicenseKey = settings.GetString("newrelic_license_key")
	}
	config := newrelic.NewConfig(newRelicApplicationName, newRelicLicenseKey)
	config.Enabled = !(newRelicApplicationName == "" || newRelicLicenseKey == "")
	if os.Getenv("NEW_RELIC_AGENT_ENABLED") != "" {
		config.Enabled, _ = strconv.ParseBool(os.Getenv("NEW_RELIC_AGENT_ENABLED"))
	}
	if !config.Enabled {
		config.AppName = "no app"
	}

	if newRelicDistributedTracingEnabled != "" {
		enabled, _ := strconv.ParseBool(newRelicDistributedTracingEnabled)

		config.CrossApplicationTracer.Enabled = !enabled
		config.DistributedTracer.Enabled = enabled
		log.WithField("config.DistributedTracer.Enabled", config.DistributedTracer.Enabled).Info("newrelic distributed tracing")

	}

	app, err := newrelic.NewApplication(config)
	if err != nil {
		log.Println(err)
		return
	}
	Newrelic = app
}

func getConfigFilePath() string {
	configFilePath := os.Getenv("APP_CONF_PATH")
	if configFilePath != "" {
		return configFilePath
	}

	configFilePath = "config"
	fs, err := os.Stat(configFilePath)
	if err == nil && fs.IsDir() {
		return configFilePath
	}

	return "conf"
}

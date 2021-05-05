package gettConfig

import (
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Settings AppConfig

type Config struct {
	*viper.Viper
}

// AppConfig ...
type AppConfig struct {
	AppEnv         string
	ServiceName    string
	Env            AppEnv
	Redis          *Config
	Rabbit         *Config
	Db             *Config
	Environment    *Config
	Secrets        *Config
	GlobalSettings *Config
	Endpoints      Endpoints
}

const configType = "yml"
const environments = "environments"
const base = "base"

var Hostname string

type AllConfig struct {
	Hostname  string
	AppConfig AppConfig
}

func getAppEnv() string {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	return appEnv
}

type EnvStore interface {
	Get(key string) string
}

type RealEnv struct{}

func (e *RealEnv) Get(key string) string {
	return os.Getenv(key)
}

func getEnvironment(configFilePath, appEnv string) *Config {
	environment := NewConfigWithEnvPrefix(configFilePath+"/"+environments, appEnv, "APP")
	return environment
}

func getGlobalSettings(configFilePath, appEnv string) *Config {
	globalSettings := NewConfigWithEnvPrefix(configFilePath+"/settings", base, "SETTINGS")
	if _, err := os.Stat(configFilePath + "/" + environments + "/" + appEnv + "." + configType); err == nil {
		// Environment config exists
		globalSettings.merge(configFilePath+"/"+environments, appEnv)
	}
	return globalSettings
}

func getEndpoints(globalSettings *Config) Endpoints {
	servicesFromConf := []string{}
	for serviceName := range globalSettings.GetStringMapString("global.endpoints") {
		serviceName = strings.Replace(serviceName, "_", "-", -1)
		servicesFromConf = append(servicesFromConf, serviceName)
	}

	for env := range globalSettings.GetStringMapStringSlice("global.env") {
		servicesFromConf = append(servicesFromConf, strings.ToUpper(env))
	}

	return newEndpoints(endpointsFromSources{
		env:  strings.Split(os.Getenv("SERVICES"), ","),
		conf: servicesFromConf,
	})
}

func ReadAllConfig(configFilePath string) *AllConfig {
	appEnv := getAppEnv()
	globalSettings := getGlobalSettings(configFilePath, appEnv)
	result := AllConfig{
		AppConfig: AppConfig{
			AppEnv:         appEnv,
			Env:            AppEnv(appEnv),
			Environment:    getEnvironment(configFilePath, appEnv),
			Db:             NewConfig(configFilePath, "dbconf"),
			Redis:          NewConfig(configFilePath, "redis"),
			Rabbit:         NewConfig(configFilePath, "rabbit"),
			Secrets:        NewConfig(configFilePath, "secrets"),
			GlobalSettings: globalSettings,
			Endpoints:      getEndpoints(globalSettings),
			ServiceName:    getServiceName(),
		},
		Hostname: getHostname(),
	}

	return &result
}

func getConfigFilePath() string {
	configFilePath := os.Getenv("APP_CONF_PATH")
	if configFilePath != "" {
		// TODO: What if the path doesn't exist or is not a directory? Should probably panic here
		return configFilePath
	}

	configFilePath = "config"
	fs, err := os.Stat(configFilePath)
	if err == nil && fs.IsDir() {
		return configFilePath
	}

	// TODO: What if "conf" doesn't exist or is not a directory? Should probably panic here
	return "conf"
}

func GetConfig() AppConfig {
	configFilePath := getConfigFilePath()
	allConfig := ReadAllConfig(configFilePath)
	Settings = allConfig.AppConfig
	Hostname = allConfig.Hostname
	return Settings
}

func getHostname() string {
	var err error
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname, err = os.Hostname()
		if err != nil {
			logrus.Error(err)
		}
	}
	return hostname
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	// will happen in local dev if no SERVICE_NAME was provided
	if serviceName == "" {
		serviceName = "myService"
	}
	return serviceName
}

const baseEnvPrefix = "CONFIG_"

func NewConfig(filepath, filename string) *Config {
	return NewConfigWithEnvPrefix(filepath, filename, strings.ToUpper(filename))
}

func NewConfigWithEnvPrefix(filepath, filename, envPrefix string) *Config {
	c := &Config{Viper: viper.New()}
	c.init(filepath, filename)
	c.SetEnvPrefix(baseEnvPrefix + envPrefix)
	c.printError(c.ReadInConfig())
	return c
}

func (c Config) init(filepath, filename string) {
	c.SetConfigType(configType)
	c.AddConfigPath(filepath)
	c.SetConfigName(filename)
	c.AutomaticEnv()
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "/", "_"))
}

func (c Config) merge(filepath, filename string) {
	c.init(filepath, filename)
	c.printError(c.MergeInConfig())
}

func (c Config) printError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func (c *Config) Get(key string) interface{} {
	val := c.Viper.Get(key)
	if strVal, ok := val.(string); ok {
		if strings.HasPrefix(strVal, "ALIAS->") {
			strVal = strings.TrimLeft(strVal, "ALIAS->")
			return c.Get(strVal)
		}
	}
	return val
}

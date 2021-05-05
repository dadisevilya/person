package gettConfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/gtforge/global_services_common_go/gett-utils/string-utils/string-format"
)

type Endpoints map[string]bool

type endpointsFromSources struct {
	env  []string
	conf []string
}

func newEndpoints(sources endpointsFromSources) Endpoints {
	res := map[string]bool{}

	source := sources.env

	if len(source) == 0 || (len(source) == 1 && source[0] == "") {
		source = sources.conf
	}

	for _, name := range source {
		name := strings.Replace(name, "_service", "", -1)
		name = strings.Replace(name, "-service", "", -1)
		res[strings.ToLower(name)] = true
	}

	return res
}

func (e *Endpoints) All() []Endpoint {
	res := []Endpoint{}
	for k := range *e {
		res = append(res, Endpoint(k))
	}
	return res
}

func (e *Endpoints) Find(name string) Endpoint {
	if !(*e)[name] {
		return ""
	}

	return Endpoint(strings.ToLower(name))
}

type Endpoint string

func (e Endpoint) Name() string {
	return string(e)
}

func (e Endpoint) HumanizedName() string {
	name := strings.Replace(string(e), "_service", "", -1)
	return stringFormat.HumanizeString(name)
}

func (e Endpoint) Hostname() string {
	if string(e) == "" {
		return ""
	}

	return e.externalHostname()
}

func (e Endpoint) InternalHostname() string {
	name := string(e)

	if name == "" {
		return ""
	}

	if os.Getenv("GET_HOSTS_FROM") == "dns" {
		return name
	}

	return e.externalHostname()
}

func (e Endpoint) IsGt() bool {
	name := strings.ToUpper(string(e))
	return name == "RU" || name == "IL" || name == "UK" || name == "US"
}

func (e Endpoint) externalHostname() string {
	name := string(e)

	if Settings.Env.IsProd() {
		return e.withProdSuffix()
	}

	if Settings.Env.IsStage() {
		return e.withStageSuffix()
	}

	if e.IsGt() {
		return Settings.GlobalSettings.GetString("global.env." + name + ".endpoints.gtforge.hostname")
	}
	name = strings.Replace(name, "-", "_", -1)
	return Settings.GlobalSettings.GetString("global.endpoints." + name + "_service.hostname")
}

const prodDomain = "gtforge.com"

func (e Endpoint) withProdSuffix() string {
	return fmt.Sprintf("%v.%v", string(e), fetchEnv("PROD_DOMAIN", prodDomain))
}

const stageDomain = "gett.io"

func (e Endpoint) withStageSuffix() string {
	name := string(e)
	var paramEnv string

	switch name {
	case "osrm":
		paramEnv = "-qa"
	case "pubnub":
		paramEnv = "-qa"
	case "relay":
		paramEnv = "scrum"
	default:
		paramEnv = "." + os.Getenv("PARAM_ENV")
	}

	return fmt.Sprintf("%v%v.%v", name, paramEnv, fetchEnv("STAGE_DOMAIN", stageDomain))
}

func fetchEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultValue
}

func HumanizedServiceName(service string) string {
	service = strings.Replace(service, "_service", "", -1)
	return stringFormat.HumanizeString(service)
}

func ServiceURLS() map[string]string {
	services := map[string]string{}
	for _, ep := range Settings.Endpoints.All() {
		services[ep.HumanizedName()] = "http://" + ep.Hostname()
	}
	return services
}

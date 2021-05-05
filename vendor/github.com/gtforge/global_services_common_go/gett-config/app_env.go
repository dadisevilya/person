package gettConfig

import "os"

type AppEnv string

func (e AppEnv) IsDev() bool {
	return e.is("development", "dev", "")
}

func (e AppEnv) IsTest() bool {
	return e.is("test")
}

func (e AppEnv) IsStage() bool {
	return e.is("stage", "staging")
}

func (e AppEnv) IsProd() bool {
	return e.is("prod", "production")
}

func (e AppEnv) is(options ...string) bool {
	env := string(e)
	for _, option := range options {
		if option == env {
			return true
		}
	}

	return false
}

func (e AppEnv) Isk8s() bool {
	return os.Getenv("GET_HOSTS_FROM") == "dns"
}

package utils

import (
	"os"

	"boilerplate/constant"
)

var Debug bool

func GetDBName() string {
	if dbName, ok := os.LookupEnv("DB_NAME"); ok {
		return dbName
	}
	return constant.Database
}
func GetClientPort() string {
	if cp, ok := os.LookupEnv("CLIENT_PORT"); ok {
		return cp
	}
	return constant.ClientPort
}

func GetAPIVersion() string {
	if av, ok := os.LookupEnv("API_VERSION"); ok {
		return av
	}
	return constant.APIVersion
}

func GetAPIGroup() string {
	if ag, ok := os.LookupEnv("API_GROUP"); ok {
		return ag
	}
	return constant.APIGroup
}

func IsDevelopment() bool {
	if Debug {
		return true
	}
	return false
}

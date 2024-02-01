package config

import (
	"flag"
	"os"
)

type serverConfig struct {
	ServerRunAddress string
	DatabaseURI      string
}

type serverConfigBuilder struct {
	serviceConfig serverConfig
}

func newServiceConfigBuilder() *serverConfigBuilder {
	return &serverConfigBuilder{
		serviceConfig: serverConfig{},
	}
}

func (sc *serverConfigBuilder) withServerRunAddress(serverRunAddress string) *serverConfigBuilder {
	sc.serviceConfig.ServerRunAddress = serverRunAddress
	return sc
}

func (sc *serverConfigBuilder) withDatabaseURI(databaseURI string) *serverConfigBuilder {
	sc.serviceConfig.DatabaseURI = databaseURI
	return sc
}

func (sc *serverConfigBuilder) build() serverConfig {
	return sc.serviceConfig
}

func BuildServer() (serverConfig, error) {
	var (
		serverRunAddress string
		databaseURI      string
	)

	flag.StringVar(&serverRunAddress, "a", "", "address:port to run server")
	flag.StringVar(&databaseURI, "d", "", "connection string for driver to establish connection to he DB")
	flag.Parse()

	if envServerRunAddress, ok := os.LookupEnv("RUN_ADDRESS"); envServerRunAddress != "" && ok {
		serverRunAddress = envServerRunAddress
	}

	if envDatabaseURI, ok := os.LookupEnv("DATABASE_URI"); envDatabaseURI != "" && ok {
		databaseURI = envDatabaseURI
	}

	return newServiceConfigBuilder().
		withServerRunAddress(serverRunAddress).
		withDatabaseURI(databaseURI).
		build(), nil
}

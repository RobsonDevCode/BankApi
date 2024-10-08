package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type AppConfig struct {
	ConnectionStrings struct {
		BankDb string `json:"bankDb"`
	}
}

var Config AppConfig

const configBasePath = "./api/Setting/Configuration/"

// SetEnvironmentSettings Sets up the environment option to be parsed into SetConfig
func SetEnvironmentSettings(environment int) string {
	switch environment {
	case 0:
		return "Development"

	case 1:
		return "UAT"

	case 2:
		return "Production"

	default:
		return ""
	}
}

// SetConfig map json configuration to our connection string struct
func SetConfig(env string) error {
	//get base config settings
	baseFile, err := openConfigFile(configBasePath+"config.json", "")
	//handle logic to close the configuration file no matter the result
	defer func() {
		if closeErr := baseFile.Close(); closeErr != nil {
			err = closeErr
			log.Fatal("Error closing config file:", closeErr)
		}
	}()
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(baseFile)

	baseConfig := AppConfig{}

	if err = decoder.Decode(&baseConfig); err != nil {
		log.Fatalf("Error Decoding Configuration File: %v", err)
		return err
	}

	envFilePath := fmt.Sprintf(configBasePath+"config.%s.json", env)

	envFile, err := openConfigFile(envFilePath, "")

	//handle logic to close the configuration file no matter the result
	defer func() {
		if closeErr := envFile.Close(); closeErr != nil {
			err = closeErr
			log.Fatal("Error closing config file:", closeErr)
		}
	}()

	decoder = json.NewDecoder(envFile)
	envConfig := AppConfig{}

	if err = decoder.Decode(&envConfig); err != nil {
		log.Fatalf("Error Decoding Environment Configuration File: %v", err)
	}

	baseConfig.ConnectionStrings = envConfig.ConnectionStrings

	Config = baseConfig
	return nil

}

// openConfigFile open correct configuration based on environment
func openConfigFile(filePath string, env string) (*os.File, error) {

	//open base configuration settings unless a specific environment is specified
	if env == "" {
		baseFile, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Error opening config file: %v", err)
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		return baseFile, nil
	} else {
		envFilePath := fmt.Sprintf(filePath, env)

		envFile, err := os.Open(envFilePath)

		if err != nil {
			log.Fatalf("Error opening %s config file: %v", env, err)
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		return envFile, nil
	}
}

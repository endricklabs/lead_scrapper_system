package config

import (
	"log"
	"os"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	DBUri string `mapstructure:"DB_URI"`

	ServerPort  string `mapstructure:"SERVER_PORT"`
	FrontendURL string `mapstructure:"FRONTEND_URL"`

	JWT struct {
		Secret string `mapstructure:"JWT_SECRET"`
		Issuer string `mapstructure:"JWT_ISSUER"`
	}

	NumberOfJobsPerRequest int64 `mapstructure:"NUMBER_OF_JOBS_PER_REQUEST"`

	NumberOfWorkersPerRequest int64 `mapstructure:"NUMBER_OF_WORKERS_PER_REQUEST"`

	LengthOfJobQueue int64 `mapstructure:"LENGTH_OF_JOB_QUEUE"`

	CheckpointNumberOfLeads int64 `mapstructure:"CHECKPOINT_NUMBER_OF_LEADS"`

	OutboxPollingInterval time.Duration `mapstructure:"OUTBOX_POLLING_INTERVAL"`
}

func LoadConfigEnv() (*Config, error) {
	if _, err := os.Stat(".env"); err != nil {
		return nil, err
	}

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	hook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	var cfg Config
	err := viper.Unmarshal(&cfg, viper.DecodeHook(hook))
	if err != nil {
		log.Println("Error while Unscrambling and Decoding Vault Path")
		return nil, err
	}

	return &cfg, nil
}

func LoadConfig() (*Config, error) {
	cfg, err := LoadConfigEnv()
	if err != nil {
		log.Fatalf("failed to load config from env: %v", err)
	}

	return cfg, nil
}

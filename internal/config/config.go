package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	defaultHttpPort        = "8000"
	defaultHttpRWTimeout   = 10 * time.Second
	defaultAppTokenTTL     = 24 * time.Hour * 90
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 24 * time.Hour * 30
	defaultLimiterRPS      = 200 //TODO: Add limiter

	defaultLogrusLevel = "trace"

	EnvLocal = "local"
	EnvProd  = "prod"
	EnvDev   = "dev"
)

type (
	Config struct {
		Environment string `mapstructure:"APP_ENV"`
		Postgres    PostgresConfig
		HTTP        HTTPConfig
		Auth        AuthConfig
		Logrus      LogrusConfig
	}

	PostgresConfig struct {
		Host     string
		Username string
		Password string
		DBName   string
	}

	HTTPConfig struct {
		Host         string
		Port         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}

	AuthConfig struct {
		JWT         JWTConfig
		AppTokenTTL time.Duration `mapstructure:"appTokenTTL"`
	}

	JWTConfig struct {
		AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
		RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
		SigningKey      string        `mapstructure:"signingKey"`
	}

	LogrusConfig struct {
		Level string
	}
)

func PrintConfig(cfg Config) {
	fmt.Printf("\tHTTP:\tHOST: %s\n", cfg.HTTP.Host)
	fmt.Printf("\tHTTP:\tPORT: %s\n", cfg.HTTP.Port)
	fmt.Printf("\tHTTP:\tR_TIMEOUT: %s\n", cfg.HTTP.ReadTimeout)
	fmt.Printf("\tHTTP:\tW_TIMEOUT: %s\n\n", cfg.HTTP.WriteTimeout)

	fmt.Printf("\tAUTH:\tApp Token TTL: %s\n", cfg.Auth.AppTokenTTL)
	fmt.Printf("\tAUTH:\tJWT:\tAccess TTL: %s\n", cfg.Auth.JWT.AccessTokenTTL)
	fmt.Printf("\tAUTH:\tJWT:\tRefresh TTL: %s\n", cfg.Auth.JWT.RefreshTokenTTL)
	fmt.Printf("\tAUTH:\tJWT:\tSigningKey: %s\n\n", cfg.Auth.JWT.SigningKey)
}

func Init(configsDir string) (*Config, error) {
	populateDefaults()

	viper.AddConfigPath(configsDir)
	viper.AddConfigPath("./")

	if err := parseConfigFile(viper.GetString("APP_ENV")); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)

	if cfg.Environment == EnvDev {
		PrintConfig(cfg)
	}

	return &cfg, nil
}

func populateDefaults() {
	viper.SetDefault("http.port", defaultHttpPort)
	viper.SetDefault("http.host", defaultHttpPort)
	viper.SetDefault("http.timeouts.read", defaultHttpRWTimeout)
	viper.SetDefault("http.timeouts.write", defaultHttpRWTimeout)
	viper.SetDefault("auth.appTokenTTL", defaultAppTokenTTL)
	viper.SetDefault("auth.accessTokenTTL", defaultAccessTokenTTL)
	viper.SetDefault("auth.refreshTokenTTL", defaultRefreshTokenTTL)
	viper.SetDefault("limiter.rps", defaultLimiterRPS)
	viper.SetDefault("logrus.level", defaultLogrusLevel)
}

func setFromEnv(cfg *Config) {
	var input string
	var is bool

	if input, is = os.LookupEnv("APP_ENV"); is {
		cfg.Environment = input
	}
	if input, is = os.LookupEnv("POSTGRES_HOST"); is {
		cfg.Postgres.Host = input
	}
	if input, is = os.LookupEnv("POSTGRES_USER"); is {
		cfg.Postgres.Username = input
	}
	if input, is = os.LookupEnv("POSTGRES_PASSWORD"); is {
		cfg.Postgres.Password = input
	}
	if input, is = os.LookupEnv("POSTGRES_DBNAME"); is {
		cfg.Postgres.DBName = input
	}
	if input, is = os.LookupEnv("JWT_SIGNING_KEY"); is {
		cfg.Auth.JWT.SigningKey = input
	}
	if input, is = os.LookupEnv("HTTP_HOST"); is {
		cfg.HTTP.Host = input
	}
}

func parseConfigFile(env string) error {
	viper.SetConfigType("yaml")
	viper.SetConfigName("main.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if env == EnvLocal {
		return nil
	}

	if env == EnvDev {
		return nil
	}

	viper.SetConfigName(env)

	return viper.MergeInConfig()
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("postgres", &cfg.Postgres); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("http", &cfg.HTTP); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("auth", &cfg.Auth); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("auth.jwt", &cfg.Auth.JWT); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("logrus", &cfg.Logrus); err != nil {
		return err
	}

	return nil
}

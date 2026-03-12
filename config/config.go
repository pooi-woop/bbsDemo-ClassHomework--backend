package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	MySQL         MySQLConfig         `mapstructure:"mysql"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Server        ServerConfig        `mapstructure:"server"`
	Logger        LoggerConfig        `mapstructure:"logger"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Email         EmailConfig         `mapstructure:"email"`
	Upload        UploadConfig        `mapstructure:"upload"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	GroupID string   `mapstructure:"group_id"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	OutputPath string `mapstructure:"output_path"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type UploadConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    int64  `mapstructure:"max_size"`
	AllowedExt string `mapstructure:"allowed_ext"`
}

type ElasticsearchConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Index    string   `mapstructure:"index"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("elasticsearch.hosts", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.index", "eyuforum")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

package database

import (
	"bbsDemo/config"
	"bbsDemo/logger"
	"fmt"
	"net"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// StartElasticsearch 启动Elasticsearch
func StartElasticsearch(cfg config.ElasticsearchConfig) (*exec.Cmd, error) {
	// 检查elasticsearch是否已经运行
	logger.Info("Checking if Elasticsearch is running...")
	_, err := net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err == nil {
		logger.Info("Elasticsearch is already running")
		return nil, nil
	}

	logger.Warn("Elasticsearch is not running, attempting to start...")
	// 启动elasticsearch
	elasticsearchCmd := exec.Command("elasticsearch")
	// 隐藏elasticsearch的输出，避免权限错误信息
	elasticsearchCmd.Stdout = nil
	elasticsearchCmd.Stderr = nil
	if err := elasticsearchCmd.Start(); err != nil {
		logger.Warn("Failed to start Elasticsearch, continuing anyway", zap.Error(err))
		return nil, err
	}

	logger.Info("Elasticsearch started successfully")
	// 等待elasticsearch启动
	time.Sleep(5 * time.Second)
	return elasticsearchCmd, nil
}

// StartKafka 启动Kafka
func StartKafka(cfg config.KafkaConfig) (*exec.Cmd, error) {
	// 检查kafka是否已经运行
	logger.Info("Checking if Kafka is running...")
	_, err := net.Dial("tcp", "localhost:9092")
	if err == nil {
		logger.Info("Kafka is already running")
		return nil, nil
	}

	logger.Warn("Kafka is not running, attempting to start...")
	// 从配置中获取kafka路径
	kafkaPath := cfg.Path
	if kafkaPath == "" {
		kafkaPath = "C:\\kafka_2.13-3.6.1" // 默认路径
	}

	kafkaServerStart := kafkaPath + "\\bin\\windows\\kafka-server-start.bat"
	kafkaConfig := kafkaPath + "\\config\\server.properties"
	kafkaCmd := exec.Command("cmd", "/c", kafkaServerStart, kafkaConfig)
	// 隐藏kafka的输出，避免权限错误信息
	kafkaCmd.Stdout = nil
	kafkaCmd.Stderr = nil
	if err := kafkaCmd.Start(); err != nil {
		logger.Warn("Failed to start Kafka, continuing anyway", zap.Error(err))
		return nil, err
	}

	logger.Info("Kafka started successfully")
	// 等待kafka启动
	time.Sleep(10 * time.Second)
	return kafkaCmd, nil
}

package database

import (
	"bbsDemo/config"
	"bbsDemo/logger"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	KafkaWriter *kafka.Writer
	KafkaReader *kafka.Reader
	kafkaCtx    = context.Background()
)

func InitKafka(cfg config.KafkaConfig) error {
	if len(cfg.Brokers) == 0 {
		return fmt.Errorf("kafka brokers not configured")
	}

	KafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	KafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	logger.Info("Kafka connected successfully",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("topic", cfg.Topic),
		zap.String("group_id", cfg.GroupID))
	return nil
}

func CloseKafka() error {
	var err error
	if KafkaWriter != nil {
		if e := KafkaWriter.Close(); e != nil {
			err = e
			logger.Error("Failed to close kafka writer", zap.Error(e))
		}
	}
	if KafkaReader != nil {
		if e := KafkaReader.Close(); e != nil {
			if err == nil {
				err = e
			}
			logger.Error("Failed to close kafka reader", zap.Error(e))
		}
	}
	if err == nil {
		logger.Info("Kafka closed successfully")
	}
	return err
}

type KafkaMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Time    int64       `json:"time"`
}

func ProduceMessage(msgType string, payload interface{}) error {
	if KafkaWriter == nil {
		return fmt.Errorf("kafka writer not initialized")
	}

	msg := KafkaMessage{
		Type:    msgType,
		Payload: payload,
		Time:    time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal kafka message", zap.Error(err))
		return err
	}

	err = KafkaWriter.WriteMessages(kafkaCtx, kafka.Message{
		Key:   []byte(msgType),
		Value: data,
	})
	if err != nil {
		logger.Error("Failed to produce kafka message",
			zap.String("type", msgType),
			zap.Error(err))
		return err
	}

	logger.Debug("Kafka message produced",
		zap.String("type", msgType))
	return nil
}

func ConsumeMessage() (*KafkaMessage, error) {
	if KafkaReader == nil {
		return nil, fmt.Errorf("kafka reader not initialized")
	}

	ctx, cancel := context.WithTimeout(kafkaCtx, time.Second)
	defer cancel()

	msg, err := KafkaReader.ReadMessage(ctx)
	if err != nil {
		// 检查错误是否是超时相关的错误
		errorStr := err.Error()
		if err == context.DeadlineExceeded || strings.Contains(strings.ToLower(errorStr), "deadline exceeded") {
			return nil, nil
		}
		logger.Error("Failed to consume kafka message", zap.Error(err))
		return nil, err
	}

	var kafkaMsg KafkaMessage
	if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
		logger.Error("Failed to unmarshal kafka message", zap.Error(err))
		return nil, err
	}

	return &kafkaMsg, nil
}

type InboxKafkaPayload struct {
	UserID string       `json:"user_id"`
	Msg    InboxMessage `json:"msg"`
}

func ProduceInboxMessage(userID int64, msg InboxMessage) error {
	if KafkaWriter == nil {
		return fmt.Errorf("kafka writer not initialized")
	}

	msg.Time = time.Now().Unix()
	payload := InboxKafkaPayload{
		UserID: strconv.FormatInt(userID, 10),
		Msg:    msg,
	}

	kafkaMsg := KafkaMessage{
		Type:    "inbox",
		Payload: payload,
		Time:    time.Now().Unix(),
	}

	data, err := json.Marshal(kafkaMsg)
	if err != nil {
		logger.Error("Failed to marshal inbox kafka message", zap.Error(err))
		return err
	}

	err = KafkaWriter.WriteMessages(kafkaCtx, kafka.Message{
		Key:   []byte("inbox"),
		Value: data,
	})
	if err != nil {
		logger.Error("Failed to produce inbox message to kafka",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return err
	}

	logger.Debug("Inbox message produced to kafka",
		zap.Int64("user_id", userID),
		zap.Any("message", msg))
	return nil
}

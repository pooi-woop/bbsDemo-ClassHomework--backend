package database

import (
	"bbsDemo/config"
	"bbsDemo/logger"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	RedisClient *redis.Client
	redisCtx    = context.Background()
)

const (
	QueueKeyEmail     = "queue:email"
	QueueKeyViewCount = "queue:view_count"
	QueueKeyLikeCount = "queue:like_count"
)

type RedisMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Time    int64       `json:"time"`
}

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type ViewCountMessage struct {
	PostID int64 `json:"post_id"`
}

type LikeCountMessage struct {
	PostID    int64  `json:"post_id,omitempty"`
	CommentID uint   `json:"comment_id,omitempty"`
	Action    string `json:"action"`
}

func InitRedis(cfg config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := RedisClient.Ping(redisCtx).Err(); err != nil {
		logger.Error("Failed to connect Redis", zap.Error(err))
		return err
	}

	logger.Info("Redis connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port))
	return nil
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

func PushMessage(queueKey string, payload interface{}) error {
	msg := RedisMessage{
		Type:    queueKey,
		Payload: payload,
		Time:    time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal message", zap.Error(err))
		return err
	}

	if err := RedisClient.LPush(redisCtx, queueKey, data).Err(); err != nil {
		logger.Error("Failed to push message to queue",
			zap.String("queue", queueKey),
			zap.Error(err))
		return err
	}

	logger.Debug("Message pushed to queue",
		zap.String("queue", queueKey),
		zap.Any("payload", payload))
	return nil
}

func PopMessage(queueKey string) (*RedisMessage, error) {
	// 使用 1 秒超时，避免永久阻塞，让 Worker 可以响应停止信号
	result, err := RedisClient.BRPop(redisCtx, time.Second, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		logger.Error("Failed to pop message from queue",
			zap.String("queue", queueKey),
			zap.Error(err))
		return nil, err
	}

	if len(result) < 2 {
		return nil, nil
	}

	var msg RedisMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		logger.Error("Failed to unmarshal message", zap.Error(err))
		return nil, err
	}

	return &msg, nil
}

func GetQueueLength(queueKey string) (int64, error) {
	return RedisClient.LLen(redisCtx, queueKey).Result()
}

func PushEmail(to, subject, body string) error {
	email := EmailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return PushMessage(QueueKeyEmail, email)
}

func PushViewCount(postID int64) error {
	view := ViewCountMessage{
		PostID: postID,
	}
	return PushMessage(QueueKeyViewCount, view)
}

func PushLikeCount(postID int64, commentID uint, action string) error {
	like := LikeCountMessage{
		PostID:    postID,
		CommentID: commentID,
		Action:    action,
	}
	return PushMessage(QueueKeyLikeCount, like)
}

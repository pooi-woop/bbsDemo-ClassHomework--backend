package queue

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Worker struct {
	emailConfig config.EmailConfig
	workerCount int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewWorker(emailConfig config.EmailConfig, workerCount int) *Worker {
	if workerCount <= 0 {
		workerCount = 3
	}
	return &Worker{
		emailConfig: emailConfig,
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}
}

func (w *Worker) Start() {
	logger.Info("Starting message queue workers", zap.Int("count", w.workerCount))

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processEmailQueue(i)
	}

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processViewCountQueue(i)
	}

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processLikeCountQueue(i)
	}

	// 启动Kafka收信箱消费者
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processInboxKafkaQueue(i)
	}
}

func (w *Worker) Stop() {
	logger.Info("Stopping message queue workers")
	close(w.stopChan)
	w.wg.Wait()
	logger.Info("All workers stopped")
}

func (w *Worker) StopWithTimeout(timeout time.Duration) bool {
	logger.Info("Stopping message queue workers with timeout", zap.Duration("timeout", timeout))
	close(w.stopChan)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All workers stopped gracefully")
		return true
	case <-time.After(timeout):
		logger.Warn("Timeout waiting for workers to stop")
		return false
	}
}

func (w *Worker) processEmailQueue(workerID int) {
	defer w.wg.Done()
	logger.Info("Email worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-w.stopChan:
			logger.Info("Email worker stopped", zap.Int("worker_id", workerID))
			return
		default:
			msg, err := database.PopMessage(database.QueueKeyEmail)
			if err != nil {
				logger.Error("Failed to pop email message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if msg == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			var email database.EmailMessage
			payloadBytes, _ := json.Marshal(msg.Payload)
			if err := json.Unmarshal(payloadBytes, &email); err != nil {
				logger.Error("Failed to unmarshal email message", zap.Error(err))
				continue
			}

			if err := w.sendEmail(email.To, email.Subject, email.Body); err != nil {
				logger.Error("Failed to send email",
					zap.String("to", email.To),
					zap.Error(err))
			} else {
				logger.Info("Email sent successfully",
					zap.String("to", email.To),
					zap.String("subject", email.Subject))
			}
		}
	}
}

func (w *Worker) processViewCountQueue(workerID int) {
	defer w.wg.Done()
	logger.Info("View count worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-w.stopChan:
			logger.Info("View count worker stopped", zap.Int("worker_id", workerID))
			return
		default:
			msg, err := database.PopMessage(database.QueueKeyViewCount)
			if err != nil {
				logger.Error("Failed to pop view count message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if msg == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			// 将 Payload 转换为 map
			payloadMap, ok := msg.Payload.(map[string]interface{})
			if !ok {
				logger.Error("Invalid payload format", zap.Any("payload", msg.Payload))
				continue
			}

			// 提取 post_id (JSON tag 使用 snake_case)
			postIDFloat, ok := payloadMap["post_id"].(float64)
			if !ok {
				logger.Error("Invalid post_id format", zap.Any("post_id", payloadMap["post_id"]))
				continue
			}
			postID := int64(postIDFloat)

			if err := database.DB.Model(&models.Post{}).
				Where("id = ?", postID).
				UpdateColumn("views", gorm.Expr("views + 1")).Error; err != nil {
				logger.Error("Failed to update view count",
					zap.Int64("post_id", postID),
					zap.Error(err))
			} else {
				logger.Info("View count updated",
					zap.Int64("post_id", postID))
			}
		}
	}
}

func (w *Worker) processLikeCountQueue(workerID int) {
	defer w.wg.Done()
	logger.Info("Like count worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-w.stopChan:
			logger.Info("Like count worker stopped", zap.Int("worker_id", workerID))
			return
		default:
			msg, err := database.PopMessage(database.QueueKeyLikeCount)
			if err != nil {
				logger.Error("Failed to pop like count message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if msg == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			var like database.LikeCountMessage
			payloadBytes, _ := json.Marshal(msg.Payload)
			if err := json.Unmarshal(payloadBytes, &like); err != nil {
				logger.Error("Failed to unmarshal like count message", zap.Error(err))
				continue
			}

			if err := w.updateLikeCount(&like); err != nil {
				logger.Error("Failed to update like count",
					zap.Int64("post_id", like.PostID),
					zap.Uint("comment_id", like.CommentID),
					zap.Error(err))
			} else {
				logger.Debug("Like count updated",
					zap.Int64("post_id", like.PostID),
					zap.Uint("comment_id", like.CommentID),
					zap.String("action", like.Action))
			}
		}
	}
}

func (w *Worker) processInboxKafkaQueue(workerID int) {
	defer w.wg.Done()
	logger.Info("Inbox Kafka worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-w.stopChan:
			logger.Info("Inbox Kafka worker stopped", zap.Int("worker_id", workerID))
			return
		default:
			msg, err := database.ConsumeMessage()
			if err != nil {
				logger.Error("Failed to consume kafka message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if msg == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			// 只处理收信箱消息
			if msg.Type != "inbox" {
				continue
			}

			payloadBytes, _ := json.Marshal(msg.Payload)
			var inboxPayload database.InboxKafkaPayload
			if err := json.Unmarshal(payloadBytes, &inboxPayload); err != nil {
				logger.Error("Failed to unmarshal inbox payload", zap.Error(err))
				continue
			}

			// 将字符串用户ID转换为int64
			userID, err := strconv.ParseInt(inboxPayload.UserID, 10, 64)
			if err != nil {
				logger.Error("Failed to parse user ID", zap.Error(err))
				continue
			}

			// 将消息存入Redis收信箱
			if err := database.PushInboxMessage(userID, inboxPayload.Msg); err != nil {
				logger.Error("Failed to push inbox message to Redis",
					zap.Int64("user_id", userID),
					zap.Error(err))
			} else {
				logger.Info("Inbox message processed",
					zap.Int64("user_id", userID),
					zap.String("type", inboxPayload.Msg.Type))
			}
		}
	}
}

func (w *Worker) updateLikeCount(like *database.LikeCountMessage) error {
	var delta int
	if like.Action == "like" {
		delta = 1
	} else if like.Action == "unlike" {
		delta = -1
	} else {
		return fmt.Errorf("invalid action: %s", like.Action)
	}

	if like.PostID > 0 {
		return database.DB.Model(&models.Post{}).
			Where("id = ?", like.PostID).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
	}

	if like.CommentID > 0 {
		return database.DB.Model(&models.Comment{}).
			Where("id = ?", like.CommentID).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
	}

	return nil
}

func (w *Worker) sendEmail(to, subject, body string) error {
	if w.emailConfig.Host == "" {
		logger.Info("Email config not set, skipping email send",
			zap.String("to", to),
			zap.String("subject", subject))
		return nil
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n%s", w.emailConfig.From, to, subject, body))

	auth := smtp.PlainAuth("", w.emailConfig.Username, w.emailConfig.Password, w.emailConfig.Host)
	addr := fmt.Sprintf("%s:%d", w.emailConfig.Host, w.emailConfig.Port)

	// 创建TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         w.emailConfig.Host,
	}

	// 连接SMTP服务器
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// 启动TLS
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// 认证
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// 设置发件人和收件人
	if err := client.Mail(w.emailConfig.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

func (w *Worker) SendInboxNotificationEmail(userID int64, email string, messageCount int, messages []database.InboxMessage) error {
	if w.emailConfig.Host == "" {
		logger.Info("Email config not set, skipping inbox notification", zap.Int64("user_id", userID))
		return nil
	}

	subject := fmt.Sprintf("您有 %d 条新消息 - EyuForum（恶雨论坛）", messageCount)

	messageList := ""
	for i, msg := range messages {
		if i >= 5 {
			messageList += fmt.Sprintf(`<p style="color: #666; font-size: 14px;">... 还有 %d 条消息，请登录查看</p>`, messageCount-5)
			break
		}
		msgType := "回复了您的帖子"
		if msg.Type == "reply_comment" {
			msgType = "回复了您的评论"
		}
		messageList += fmt.Sprintf(`
<div style="background: #f8f9fa; border-left: 3px solid #667eea; padding: 12px; margin-bottom: 10px; border-radius: 4px;">
    <p style="margin: 0; color: #333; font-size: 14px;">
        <strong>用户 %d</strong> %s
    </p>
</div>`, msg.SenderID, msgType)
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f4;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo {
            font-size: 28px;
            font-weight: bold;
            color: #4a90e2;
        }
        .notification-box {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #ffffff;
            font-size: 24px;
            font-weight: bold;
            padding: 20px;
            text-align: center;
            border-radius: 8px;
            margin: 20px 0;
            box-shadow: 0 4px 6px rgba(102, 126, 234, 0.3);
        }
        .message-list {
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            color: #999;
            font-size: 14px;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
        }
        .btn {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #ffffff;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 6px;
            font-weight: bold;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">📬 EyuForum（恶雨论坛）</div>
        </div>
        
        <div class="notification-box">
            您有 %d 条新消息！
        </div>
        
        <div class="message-list">
            %s
        </div>
        
        <div style="text-align: center;">
            <a href="http://localhost:5173" class="btn">立即查看</a>
        </div>
        
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2026 EyuForum（恶雨论坛）. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, messageCount, messageList)

	return w.sendEmail(email, subject, body)
}

/*消息队列整体架构如下
┌─────────────────────────────────────────┐
│              Worker Pool                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │ Email   │ │ View    │ │ Like    │   │
│  │ Worker  │ │ Worker  │ │ Worker  │   │
│  │  (N个)  │ │  (N个)  │ │  (N个)  │   │
│  └────┬────┘ └────┬────┘ └────┬────┘   │
│       └─────────────┴─────────────┘      │
│              3 个独立队列                 │
│    QueueKeyEmail | QueueKeyViewCount     │
│              QueueKeyLikeCount           │
└─────────────────────────────────────────┘
*/

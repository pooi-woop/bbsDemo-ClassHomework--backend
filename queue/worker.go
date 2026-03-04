package queue

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"encoding/json"
	"fmt"
	"net/smtp"
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
}

func (w *Worker) Stop() {
	logger.Info("Stopping message queue workers")
	close(w.stopChan)
	w.wg.Wait()
	logger.Info("All workers stopped")
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

			var view database.ViewCountMessage
			payloadBytes, _ := json.Marshal(msg.Payload)
			if err := json.Unmarshal(payloadBytes, &view); err != nil {
				logger.Error("Failed to unmarshal view count message", zap.Error(err))
				continue
			}

			if err := database.DB.Model(&models.Post{}).
				Where("id = ?", view.PostID).
				UpdateColumn("views", gorm.Expr("views + 1")).Error; err != nil {
				logger.Error("Failed to update view count",
					zap.Uint("post_id", view.PostID),
					zap.Error(err))
			} else {
				logger.Debug("View count updated",
					zap.Uint("post_id", view.PostID))
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
					zap.Uint("post_id", like.PostID),
					zap.Uint("comment_id", like.CommentID),
					zap.Error(err))
			} else {
				logger.Debug("Like count updated",
					zap.Uint("post_id", like.PostID),
					zap.Uint("comment_id", like.CommentID),
					zap.String("action", like.Action))
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

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))

	auth := smtp.PlainAuth("", w.emailConfig.Username, w.emailConfig.Password, w.emailConfig.Host)
	addr := fmt.Sprintf("%s:%d", w.emailConfig.Host, w.emailConfig.Port)

	return smtp.SendMail(addr, auth, w.emailConfig.From, []string{to}, msg)
}

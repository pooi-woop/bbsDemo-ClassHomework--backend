package scheduler

import (
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"bbsDemo/queue"
	"context"
	"time"

	"go.uber.org/zap"
)

type Scheduler struct {
	worker   *queue.Worker
	stopChan chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewScheduler(worker *queue.Worker) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		worker:   worker,
		stopChan: make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Scheduler) Start() {
	logger.Info("Starting scheduler")

	go s.dailyInboxNotification()
}

func (s *Scheduler) Stop() {
	logger.Info("Stopping scheduler")
	s.cancel()
	close(s.stopChan)
	logger.Info("Scheduler stopped")
}

func (s *Scheduler) dailyInboxNotification() {
	for {
		select {
		case <-s.stopChan:
			return
		case <-s.ctx.Done():
			return
		default:
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())
			duration := next.Sub(now)

			logger.Info("Next inbox notification scheduled",
				zap.Time("next_run", next),
				zap.Duration("duration", duration))

			select {
			case <-time.After(duration):
				s.sendInboxNotifications()
			case <-s.stopChan:
				return
			case <-s.ctx.Done():
				return
			}
		}
	}
}

func (s *Scheduler) sendInboxNotifications() {
	logger.Info("Starting daily inbox notification")

	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		logger.Error("Failed to fetch users", zap.Error(err))
		return
	}

	successCount := 0
	for _, user := range users {
		messages, total, err := database.GetInboxMessages(user.ID, 1, 10)
		if err != nil {
			logger.Error("Failed to get inbox messages",
				zap.Int64("user_id", user.ID),
				zap.Error(err))
			continue
		}

		if total > 0 {
			if err := s.worker.SendInboxNotificationEmail(user.ID, user.Email, int(total), messages); err != nil {
				logger.Error("Failed to send inbox notification email",
					zap.Int64("user_id", user.ID),
					zap.String("email", user.Email),
					zap.Error(err))
			} else {
				successCount++
				logger.Info("Inbox notification email sent",
					zap.Int64("user_id", user.ID),
					zap.String("email", user.Email),
					zap.Int64("message_count", total))
			}
		}
	}

	logger.Info("Daily inbox notification completed",
		zap.Int("total_users", len(users)),
		zap.Int("notified_users", successCount))
}

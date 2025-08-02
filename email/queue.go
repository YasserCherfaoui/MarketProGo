package email

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/redis/go-redis/v9"
)

// EmailQueue interface defines the contract for email queue operations
type EmailQueue interface {
	Enqueue(email *models.Email) error
	Dequeue() (*models.Email, error)
	MarkAsProcessed(emailID string) error
	MarkAsFailed(emailID string, error string) error
	GetFailedEmails() ([]*models.Email, error)
	GetQueueSize() (int64, error)
	ClearQueue() error
}

// MockEmailQueue implements EmailQueue for testing and development
type MockEmailQueue struct {
	emails []*models.Email
}

// NewMockEmailQueue creates a new mock email queue
func NewMockEmailQueue() *MockEmailQueue {
	return &MockEmailQueue{
		emails: make([]*models.Email, 0),
	}
}

// Enqueue adds an email to the mock queue
func (m *MockEmailQueue) Enqueue(email *models.Email) error {
	m.emails = append(m.emails, email)
	fmt.Printf("MOCK QUEUE: Enqueued email to %s with subject: %s\n",
		email.Recipients[0].Email, email.Subject)
	return nil
}

// Dequeue removes and returns the next email from the mock queue
func (m *MockEmailQueue) Dequeue() (*models.Email, error) {
	if len(m.emails) == 0 {
		return nil, nil
	}

	email := m.emails[0]
	m.emails = m.emails[1:]
	return email, nil
}

// MarkAsProcessed marks an email as successfully processed
func (m *MockEmailQueue) MarkAsProcessed(emailID string) error {
	fmt.Printf("MOCK QUEUE: Marked email %s as processed\n", emailID)
	return nil
}

// MarkAsFailed marks an email as failed
func (m *MockEmailQueue) MarkAsFailed(emailID string, errorMsg string) error {
	fmt.Printf("MOCK QUEUE: Marked email %s as failed: %s\n", emailID, errorMsg)
	return nil
}

// GetFailedEmails returns failed emails from the mock queue
func (m *MockEmailQueue) GetFailedEmails() ([]*models.Email, error) {
	return []*models.Email{}, nil
}

// GetQueueSize returns the size of the mock queue
func (m *MockEmailQueue) GetQueueSize() (int64, error) {
	return int64(len(m.emails)), nil
}

// ClearQueue clears the mock queue
func (m *MockEmailQueue) ClearQueue() error {
	m.emails = make([]*models.Email, 0)
	return nil
}

// RedisEmailQueue implements EmailQueue using Redis
type RedisEmailQueue struct {
	client *redis.Client
	queue  string
}

// NewRedisEmailQueue creates a new Redis email queue
func NewRedisEmailQueue(client *redis.Client, queueName string) *RedisEmailQueue {
	return &RedisEmailQueue{
		client: client,
		queue:  queueName,
	}
}

// Enqueue adds an email to the queue
func (r *RedisEmailQueue) Enqueue(email *models.Email) error {
	// Serialize email to JSON
	emailData, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	// Add to Redis list (left push for FIFO)
	ctx := context.Background()
	err = r.client.LPush(ctx, r.queue, emailData).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue email: %w", err)
	}

	return nil
}

// Dequeue removes and returns the next email from the queue
func (r *RedisEmailQueue) Dequeue() (*models.Email, error) {
	ctx := context.Background()

	// Pop from right side of list (FIFO)
	result, err := r.client.BRPop(ctx, 5*time.Second, r.queue).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Queue is empty
		}
		return nil, fmt.Errorf("failed to dequeue email: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	// Deserialize email from JSON
	var email models.Email
	err = json.Unmarshal([]byte(result[1]), &email)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal email: %w", err)
	}

	return &email, nil
}

// MarkAsProcessed marks an email as successfully processed
func (r *RedisEmailQueue) MarkAsProcessed(emailID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("email:processed:%s", emailID)

	// Store processed email with timestamp
	data := map[string]interface{}{
		"processed_at": time.Now().Unix(),
		"status":       "processed",
	}

	processedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal processed data: %w", err)
	}

	err = r.client.Set(ctx, key, processedData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to mark email as processed: %w", err)
	}

	return nil
}

// MarkAsFailed marks an email as failed
func (r *RedisEmailQueue) MarkAsFailed(emailID string, errorMsg string) error {
	ctx := context.Background()
	key := fmt.Sprintf("email:failed:%s", emailID)

	// Store failed email with timestamp and error
	data := map[string]interface{}{
		"failed_at": time.Now().Unix(),
		"status":    "failed",
		"error":     errorMsg,
	}

	failedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal failed data: %w", err)
	}

	err = r.client.Set(ctx, key, failedData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to mark email as failed: %w", err)
	}

	return nil
}

// GetFailedEmails returns failed emails from Redis
func (r *RedisEmailQueue) GetFailedEmails() ([]*models.Email, error) {
	ctx := context.Background()
	pattern := "email:failed:*"

	// Get all failed email keys
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get failed email keys: %w", err)
	}

	var failedEmails []*models.Email
	for _, key := range keys {
		// Extract email ID from key
		_ = strings.TrimPrefix(key, "email:failed:")

		// Get email data from database (you might want to store the full email data)
		// For now, we'll return empty list as this is a simplified implementation
	}

	return failedEmails, nil
}

// GetQueueSize returns the size of the Redis queue
func (r *RedisEmailQueue) GetQueueSize() (int64, error) {
	ctx := context.Background()
	size, err := r.client.LLen(ctx, r.queue).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}
	return size, nil
}

// ClearQueue clears the Redis queue
func (r *RedisEmailQueue) ClearQueue() error {
	ctx := context.Background()
	err := r.client.Del(ctx, r.queue).Err()
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}
	return nil
}

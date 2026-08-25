package service

import (
	"context"
	"time"

	"flamingo/app/conversation/internal/model"

	"gorm.io/gorm"
)

// ConversationRepository is the conversation repository interface
type ConversationRepository interface {
	// Upsert creates or updates a conversation
	Upsert(ctx context.Context, conversation *model.Conversation) error
	// GetByID retrieves a conversation by conversation ID
	GetByID(ctx context.Context, conversationID string) (*model.Conversation, error)
	// GetByUserAndTarget retrieves a conversation by user ID and target ID
	GetByUserAndTarget(ctx context.Context, userID string, conversationType model.ConversationType, targetID string) (*model.Conversation, error)
	// ListByUser retrieves the user's conversation list
	ListByUser(ctx context.Context, userID string, limit int, updatedBefore *time.Time) ([]*model.Conversation, error)
	// Delete deletes a conversation
	Delete(ctx context.Context, userID, conversationID string) error
	// SetPinned sets pinned status
	SetPinned(ctx context.Context, userID, conversationID string, pinned bool, pinTime *time.Time) error
	// SetMuted sets muted status
	SetMuted(ctx context.Context, userID, conversationID string, muted bool) error
	// SetBurnAfterReading sets burn after reading duration
	SetBurnAfterReading(ctx context.Context, userID, conversationID string, duration int32) error
	// SetAutoDelete sets auto delete duration
	SetAutoDelete(ctx context.Context, userID, conversationID string, duration int32) error
	// ClearUnread clears unread count
	ClearUnread(ctx context.Context, userID, conversationID string) error
	// IncrUnread increments unread count
	IncrUnread(ctx context.Context, userID, conversationID string, count int32) error
	// SumUnread counts user's total unread count
	SumUnread(ctx context.Context, userID string) (int32, error)
	// WithTx uses transaction
	WithTx(tx *gorm.DB) ConversationRepository
}

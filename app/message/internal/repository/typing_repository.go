package repository

import (
	"context"
	"fmt"
	"time"

	"flamingo/app/message/internal/service"
	"flamingo/pkg/cache"

	"github.com/go-kratos/kratos/v2/log"
)

type typingRepositoryImpl struct {
	cache cache.Cache
	log   *log.Helper
}

// NewTypingRepository creates typing status cache repository
func NewTypingRepository(cache cache.Cache, logger log.Logger) service.TypingRepository {
	return &typingRepositoryImpl{
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

func (r *typingRepositoryImpl) SetState(ctx context.Context, conversationID, fromUserID string, ttl time.Duration) error {
	return r.cache.Set(ctx, r.stateKey(conversationID, fromUserID), "1", ttl)
}

func (r *typingRepositoryImpl) ClearState(ctx context.Context, conversationID, fromUserID string) error {
	return r.cache.Del(ctx, r.stateKey(conversationID, fromUserID), r.emitKey(conversationID, fromUserID))
}

func (r *typingRepositoryImpl) AcquireEmitToken(ctx context.Context, conversationID, fromUserID string, ttl time.Duration) (bool, error) {
	return r.cache.SetNX(ctx, r.emitKey(conversationID, fromUserID), "1", ttl)
}

func (r *typingRepositoryImpl) stateKey(conversationID, fromUserID string) string {
	return fmt.Sprintf("msg:typing:state:%s:%s", conversationID, fromUserID)
}

func (r *typingRepositoryImpl) emitKey(conversationID, fromUserID string) string {
	return fmt.Sprintf("msg:typing:emit:%s:%s", conversationID, fromUserID)
}

package repository

import (
	"context"

	"flamingo/app/rtc/internal/model"
	"flamingo/app/rtc/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type meetingRepository struct {
	db  *gorm.DB
	log *log.Helper
}

func NewMeetingRepository(db *gorm.DB, logger log.Logger) service.MeetingRepository {
	return &meetingRepository{
		db:  db,
		log: log.NewHelper(logger),
	}
}

func (r *meetingRepository) CreateMeeting(ctx context.Context, meeting *model.MeetingRoom) error {
	return r.db.WithContext(ctx).Create(meeting).Error
}

func (r *meetingRepository) GetMeetingByRoomID(ctx context.Context, roomID string) (*model.MeetingRoom, error) {
	var meeting model.MeetingRoom
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (r *meetingRepository) UpdateMeeting(ctx context.Context, meeting *model.MeetingRoom) error {
	return r.db.WithContext(ctx).Save(meeting).Error
}

func (r *meetingRepository) ListActiveMeetings(ctx context.Context, page, pageSize int) ([]*model.MeetingRoom, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var meetings []*model.MeetingRoom
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MeetingRoom{}).
		Where("status = ?", model.MeetingStatusActive)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&meetings).Error; err != nil {
		return nil, 0, err
	}
	return meetings, total, nil
}

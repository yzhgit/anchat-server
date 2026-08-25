package repository

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewGroupJoinRequestRepository,
	NewGroupMemberRepository,
	NewGroupPinnedMessageRepository,
	NewGroupQRCodeRepository,
	NewGroupRepository,
	NewGroupSettingRepository,
)

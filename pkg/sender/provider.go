package sender

import (
	"flamingo/pkg/config"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.FieldsOf(new(config.Verify), "Email"),
	NewEmailSender,
	NewNoopSMSSender,
)

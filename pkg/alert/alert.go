package alert

import "github.com/mike1808/ax/pkg/backend/common"

type Alerter interface {
	SendAlert(lm common.LogMessage) error
}

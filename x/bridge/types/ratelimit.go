package types

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
)

func DefaultEpochInfo() EpochInfo {
	return EpochInfo{
		StartTime:               time.Time{},
		Duration:                time.Hour * 24,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}
}

func (epoch *EpochInfo) Validate() error {
	if epoch.Duration == 0 {
		return ErrEpochDurationZero
	}
	if epoch.CurrentEpochStartHeight < 0 {
		return errors.New("epoch CurrentEpochStartHeight must be non-negative")
	}
	return nil
}

// Adds an amount to the rate limit's flow after a packet was sent
// Returns an error if the new inflow will cause the rate limit to exceed its quota
func (r *RateLimit) CheckAddInflowThreshold(amount math.Int) error {
	netInflow := r.CurrentInflow.Add(amount)
	if netInflow.GT(r.Ratelimit) {
		return errorsmod.Wrap(ErrInflowThresholdExceeded,
			fmt.Sprintf("tx amount (%v) exceeds threshold (%v)", amount.String(), r.Ratelimit.String()),
		)
	}

	r.CurrentInflow = netInflow
	return nil
}

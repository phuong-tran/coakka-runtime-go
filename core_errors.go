package coakka_v2_connector

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ErrClosed means the Go owner has refused new calls during Close. It does not
// reconstruct a native lifecycle state or replace a Core-issued status.
var ErrClosed = errors.New("runtime client closed")

// CoreStatusError preserves an unsuccessful Core operation's numeric status,
// including future unknown values. Inspect with errors.As, not message text.
type CoreStatusError struct {
	Operation string
	Status    int
}

func (e *CoreStatusError) Error() string {
	return fmt.Sprintf("%s failed status=%d", e.Operation, e.Status)
}

// CoreContractError reports a boundary invariant violation, not a guessed Core
// causal status. No native result or effective setting is fabricated.
type CoreContractError struct {
	Operation string
	Detail    string
}

func (e *CoreContractError) Error() string {
	return fmt.Sprintf("%s: Core contract violation: %s", e.Operation, e.Detail)
}

// FeatureUnavailableError means Core successfully reported the feature absent.
// Failure to inspect Core is instead returned as CoreStatusError.
type FeatureUnavailableError struct{ Feature string }

func (e *FeatureUnavailableError) Error() string {
	return fmt.Sprintf("Core does not advertise %s", e.Feature)
}

// C-string representation is a connector responsibility. Empty strings remain
// unchanged for Core validation; only unrepresentable input is refused here.
func validateNativeText(fields ...[2]string) error {
	for _, field := range fields {
		if len(field[1]) > math.MaxInt32 || strings.IndexByte(field[1], 0) >= 0 {
			return fmt.Errorf("%s cannot cross the C-string boundary", field[0])
		}
	}
	return nil
}

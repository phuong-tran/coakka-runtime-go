package coakka_v2_connector

import (
	"encoding/json"
	"errors"
	"fmt"
)

const runtimeFeatureLaneOwnerGrants uint32 = 1 << 25

// LaneOwnerConfig identifies and advertises one exact receiver or publisher replica.
// AdvertisedHost must route directly to that owner, never through a load-balancing service.
type LaneOwnerConfig struct {
	OwnerInstanceID string
	AdvertisedHost  string
}

func (c LaneOwnerConfig) validate() error {
	if err := validateVisibleASCII("owner instance ID", c.OwnerInstanceID, 127); err != nil {
		return err
	}
	return validateVisibleASCII("advertised host", c.AdvertisedHost, 255)
}

func validateVisibleASCII(name, value string, max int) error {
	if len(value) == 0 || len(value) > max {
		return fmt.Errorf("%s must contain 1..%d visible ASCII bytes", name, max)
	}
	for i := range len(value) {
		if value[i] < 0x21 || value[i] > 0x7e {
			return fmt.Errorf("%s must contain only visible ASCII bytes", name)
		}
	}
	return nil
}

// LaneOwnerEndpoint is the exact owner endpoint projected into a grant.
type LaneOwnerEndpoint struct {
	OwnerInstanceID string
	AdvertisedHost  string
	Port            uint16
}

func (e LaneOwnerEndpoint) validate() error {
	if err := validateVisibleASCII("owner instance ID", e.OwnerInstanceID, 127); err != nil {
		return err
	}
	if err := validateVisibleASCII("advertised host", e.AdvertisedHost, 255); err != nil {
		return err
	}
	if e.Port == 0 {
		return errors.New("owner port must be non-zero")
	}
	return nil
}

// UnmarshalJSON validates a File grant received over a trusted control plane.
func (g *FileReceiveGrant) UnmarshalJSON(data []byte) error {
	type wireGrant FileReceiveGrant
	var value wireGrant
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := value.Owner.validate(); err != nil {
		return err
	}
	if err := validateVisibleASCII("transfer ID", value.TransferID, 64); err != nil {
		return err
	}
	if err := validateVisibleASCII("authorization token", value.AuthorizationToken, 128); err != nil {
		return err
	}
	*g = FileReceiveGrant(value)
	return nil
}

// UnmarshalJSON validates a Stream grant received over a trusted control plane.
func (g *StreamPublishGrant) UnmarshalJSON(data []byte) error {
	type wireGrant StreamPublishGrant
	var value wireGrant
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := value.Owner.validate(); err != nil {
		return err
	}
	if err := validateVisibleASCII("session ID", value.SessionID, 64); err != nil {
		return err
	}
	if err := validateVisibleASCII("authorization token", value.AuthorizationToken, 128); err != nil {
		return err
	}
	if value.FormatID == 0 || value.MaxFrameBytes == 0 || value.MaxFrameBytes > 4*1024*1024 {
		return errors.New("stream grant format and frame bounds must be non-zero and valid")
	}
	*g = StreamPublishGrant(value)
	return nil
}

func requireLaneOwnerGrants(bindings *nativeBindings) error {
	info, err := bindings.readRuntimeInfo()
	if err != nil {
		return err
	}
	if info.FeatureFlags&runtimeFeatureLaneOwnerGrants == 0 {
		return errors.New("native runtime does not advertise lane-owner grants")
	}
	return nil
}

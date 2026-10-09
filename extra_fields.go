package frontal

import (
	"encoding/json"
	"fmt"
)

// ExtraFields stores response properties that a generated or caller-owned
// response type does not model.
type ExtraFields map[string]json.RawMessage

// DecodeWithExtraFields decodes data into target and returns properties not
// listed in knownFields as raw JSON for forward-compatible handling.
func DecodeWithExtraFields(data []byte, target any, knownFields ...string) (ExtraFields, error) {
	if target == nil {
		return nil, fmt.Errorf("frontal: response target is nil")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return nil, fmt.Errorf("frontal: decode response: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("frontal: decode response properties: %w", err)
	}
	for _, name := range knownFields {
		delete(fields, name)
	}
	return ExtraFields(fields), nil
}

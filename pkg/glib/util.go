package glib

import (
	"encoding/json"
	"fmt"
)

func convertToStruct(data json.Marshaler, dst interface{}) error {
	encoded, err := data.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %v", err)
	}

	if err := json.Unmarshal(encoded, dst); err != nil {
		return fmt.Errorf("failed to decode as JSON: %v", err)
	}

	return nil
}

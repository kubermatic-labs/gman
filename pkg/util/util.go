package util

import (
	"encoding/json"
	"fmt"
)

func StringSliceContains(s []string, needle string) bool {
	for _, item := range s {
		if item == needle {
			return true
		}
	}

	return false
}

func ConvertToStruct(data json.Marshaler, dst interface{}) error {
	encoded, err := data.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %v", err)
	}

	if err := json.Unmarshal(encoded, dst); err != nil {
		return fmt.Errorf("failed to decode as JSON: %v", err)
	}

	return nil
}

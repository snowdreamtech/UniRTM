// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"testing"
)

func TestConfig_Unmarshal(t *testing.T) {
	var sa StringArray

	err := json.Unmarshal([]byte(`"value"`), &sa)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = json.Unmarshal([]byte(`["val1", "val2"]`), &sa)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

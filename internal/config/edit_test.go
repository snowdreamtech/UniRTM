// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"strings"
	"testing"
)

func TestUpsertEnvVar(t *testing.T) {
	// 1. Empty content
	content := ""
	res := UpsertEnvVar(content, "FOO", "bar")
	if !strings.Contains(res, "[env]") || !strings.Contains(res, "FOO = \"bar\"") {
		t.Errorf("UpsertEnvVar failed on empty content: %s", res)
	}

	// 2. Existing [env] but no FOO
	content = "[env]\nBAZ = \"qux\"\n[tools]\nnode = \"18\""
	res = UpsertEnvVar(content, "FOO", "bar")
	if !strings.Contains(res, "FOO = \"bar\"") {
		t.Errorf("UpsertEnvVar failed to add to existing [env]: %s", res)
	}
	// FOO should be before [tools]
	if strings.Index(res, "FOO") > strings.Index(res, "[tools]") {
		t.Errorf("UpsertEnvVar added key after next section: %s", res)
	}

	// 3. Existing [env] with FOO
	content = "[env]\nFOO = \"old\"\n[tools]\nnode = \"18\""
	res = UpsertEnvVar(content, "FOO", "new")
	if !strings.Contains(res, "FOO = \"new\"") || strings.Contains(res, "FOO = \"old\"") {
		t.Errorf("UpsertEnvVar failed to update existing key: %s", res)
	}
}

func TestUnsetEnvVar(t *testing.T) {
	// 1. Key exists
	content := "[env]\nFOO = \"bar\"\nBAZ = \"qux\"\n"
	res, changed := UnsetEnvVar(content, "FOO")
	if !changed {
		t.Errorf("UnsetEnvVar should return changed=true")
	}
	if strings.Contains(res, "FOO") {
		t.Errorf("UnsetEnvVar failed to remove key: %s", res)
	}

	// 2. Key does not exist
	res, changed = UnsetEnvVar(content, "NONEXISTENT")
	if changed {
		t.Errorf("UnsetEnvVar should return changed=false for nonexistent key")
	}
}

func TestUpsertToolVersion(t *testing.T) {
	// 1. Empty content
	content := ""
	res := UpsertToolVersion(content, "node", "18")
	if !strings.Contains(res, "[tools]") || !strings.Contains(res, "node = \"18\"") {
		t.Errorf("UpsertToolVersion failed on empty content: %s", res)
	}

	// 2. Existing [tools]
	content = "[tools]\npython = \"3.10\"\n[env]\nFOO = \"bar\""
	res = UpsertToolVersion(content, "node", "18")
	if !strings.Contains(res, "node = \"18\"") {
		t.Errorf("UpsertToolVersion failed to add to existing [tools]: %s", res)
	}
	if strings.Index(res, "node = \"18\"") > strings.Index(res, "[env]") {
		t.Errorf("UpsertToolVersion added key after next section: %s", res)
	}

	// 3. Update existing tool
	content = "[tools]\nnode = \"16\"\npython = \"3.10\""
	res = UpsertToolVersion(content, "node", "18")
	if !strings.Contains(res, "node = \"18\"") || strings.Contains(res, "node = \"16\"") {
		t.Errorf("UpsertToolVersion failed to update existing tool: %s", res)
	}
}

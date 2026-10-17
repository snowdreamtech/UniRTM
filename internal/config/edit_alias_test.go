package config

import (
	"strings"
	"testing"
)

func TestUpsertAlias_Empty(t *testing.T) {
	content := ""
	res := UpsertAlias(content, "node", "lts", "20")
	if !strings.Contains(res, "[aliases.node]") || !strings.Contains(res, `lts = "20"`) {
		t.Fatalf("UpsertAlias on empty failed: %s", res)
	}
}

func TestUpsertAlias_MultipleAndOverwrite(t *testing.T) {
	content := `[settings]
jobs = 4

[aliases.node]
lts = "18"
`
	// 1. Update existing alias
	updated := UpsertAlias(content, "node", "lts", "20")
	if strings.Contains(updated, `lts = "18"`) || !strings.Contains(updated, `lts = "20"`) {
		t.Fatalf("UpsertAlias overwrite failed: %s", updated)
	}

	// 2. Append alias to existing tool
	appended := UpsertAlias(updated, "node", "latest", "22")
	if !strings.Contains(appended, `lts = "20"`) || !strings.Contains(appended, `latest = "22"`) {
		t.Fatalf("UpsertAlias append failed: %s", appended)
	}

	// 3. Add alias for a new tool
	newTool := UpsertAlias(appended, "python", "stable", "3.12")
	if !strings.Contains(newTool, "[aliases.python]") || !strings.Contains(newTool, `stable = "3.12"`) {
		t.Fatalf("UpsertAlias new tool failed: %s", newTool)
	}

	// Verify parser
	aliases := GetAliasesFromContent(newTool)
	if aliases["node"]["lts"] != "20" || aliases["node"]["latest"] != "22" || aliases["python"]["stable"] != "3.12" {
		t.Fatalf("GetAliasesFromContent returned unexpected: %+v", aliases)
	}
}

func TestRemoveAlias(t *testing.T) {
	content := `[aliases.node]
lts = "20"
latest = "22"

[aliases.python]
stable = "3.12"
`
	// 1. Remove one alias from multi-entry section
	res, ok := RemoveAlias(content, "node", "lts")
	if !ok || strings.Contains(res, `lts = "20"`) || !strings.Contains(res, `latest = "22"`) {
		t.Fatalf("RemoveAlias single entry failed: %s", res)
	}

	// 2. Remove last remaining alias for python -> section header should be cleaned up
	res2, ok2 := RemoveAlias(res, "python", "stable")
	if !ok2 || strings.Contains(res2, "[aliases.python]") || strings.Contains(res2, `stable = "3.12"`) {
		t.Fatalf("RemoveAlias cleanup section failed: %s", res2)
	}

	// 3. Remove non-existent alias
	_, ok3 := RemoveAlias(res2, "nonexistent", "foo")
	if ok3 {
		t.Fatalf("RemoveAlias on non-existent returned true")
	}
}

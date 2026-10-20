// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
)

// GetAllBuiltinRules returns instances of all Go-native built-in guard rules
func GetAllBuiltinRules() []Rule {
	return []Rule{
		NewTrailingWhitespaceRule(),
		NewEndOfFileRule(),
		NewUtf8BOMRule(),
		NewMergeConflictRule(),
		NewPrivateKeyRule(),
		NewEnvSafetyRule(),
		NewLargeFilesRule(),
		NewCaseConflictRule(),
		NewLockfileCoexistenceRule(),
		NewLockfileConflictsRule(),
		NewForbiddenFilesRule(),
		NewSymlinksRule(),
		NewExecutablesHaveShebangsRule(),
		NewShebangScriptsAreExecutableRule(),
		NewSvgSafetyRule(),
		NewJsonRule(),
		NewTomlRule(),
		NewMixedLineEndingsRule(),
		NewXmlRule(),
	}
}


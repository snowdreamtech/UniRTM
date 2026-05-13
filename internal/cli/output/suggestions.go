// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"fmt"
	"io"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Suggest prints fuzzy suggestions for a target string from a list of candidates.
func Suggest(w io.Writer, target string, candidates []string) {
	if len(candidates) == 0 || len(target) < 2 {
		return
	}

	type match struct {
		target   string
		distance int
	}
	var matches []match

	for _, c := range candidates {
		dist := fuzzy.LevenshteinDistance(target, c)
		// Suggest if distance is small (e.g. <= 2 or <= 30% of length)
		if dist <= 2 || dist <= len(target)/2 {
			matches = append(matches, match{target: c, distance: dist})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].distance < matches[j].distance
	})

	if len(matches) > 0 {
		fmt.Fprintf(w, "\nDid you mean one of these?\n")
		limit := 3
		if len(matches) < limit {
			limit = len(matches)
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(w, "  - %s\n", matches[i].target)
		}
	}
}

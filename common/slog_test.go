// Package slog implements the third iteration towards a logging framework that makes our CTO happy
// The current idea is that the Golang slog package is great, we only need some cosmetics to always log some
// values from the context
package common

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestShortenFile(t *testing.T) {
	testCases := []struct {
		test     string
		expected string
	}{
		{test: "/Users/totomz/Library/services/pacioli/whatever.go", expected: "services/pacioli/whatever.go"},
		{test: "/Users/totomz/Library/Caches/JetBrains/IntelliJIdea2024", expected: "/Users/totomz/Library/Caches/JetBrains/IntelliJIdea2024"},
		{test: "pippo.pluto/paperino blabla", expected: "pippo.pluto/paperino blabla"},
	}

	for _, testCase := range testCases {
		got := shortenFilePath(testCase.test)
		if diff := cmp.Diff(testCase.expected, got); len(diff) > 0 {
			t.Errorf("\ninvalid uniqe words count\n%s", diff)
		}
	}
}

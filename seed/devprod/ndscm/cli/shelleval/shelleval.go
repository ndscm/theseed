package shelleval

import _ "embed"

//go:embed nd_snippet.sh
var ndSnippet string

//go:embed nd_completions_snippet.sh
var ndCompletionsSnippet string

func NdSnippet() string {
	return ndSnippet
}

func NdCompletionsSnippet() string {
	return ndCompletionsSnippet
}

package configloader

type BazelWatcherRule struct {
	Build StringOrStrings `json:"build"`
	Run   StringOrStrings `json:"run"`
}

type WatchGenerateRule struct {
	Target StringOrStrings `json:"target,omitempty"`

	WatchRepo StringOrStrings `json:"watchRepo,omitempty"`

	Watch StringOrStrings `json:"watch"`

	Run StringOrStrings `json:"run,omitempty"`

	Bazel *BazelWatcherRule `json:"bazel,omitempty"`
}

type DirConfig struct {
	Format    map[string]WatchGenerateRule `json:"format,omitempty"`
	Vendor    map[string]WatchGenerateRule `json:"vendor,omitempty"`
	Bootstrap map[string]WatchGenerateRule `json:"bootstrap,omitempty"`
	Tidy      map[string]WatchGenerateRule `json:"tidy,omitempty"`
	Lock      map[string]WatchGenerateRule `json:"lock,omitempty"`
	Build     map[string]WatchGenerateRule `json:"build,omitempty"`
	Test      map[string]WatchGenerateRule `json:"test,omitempty"`
}

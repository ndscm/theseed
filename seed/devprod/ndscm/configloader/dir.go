package configloader

type WatchGenerateRule struct {
	Target StringOrStrings `json:"target,omitempty"`

	WatchRepo StringOrStrings `json:"watchRepo,omitempty"`

	Watch StringOrStrings `json:"watch"`

	NeedBazelBuild StringOrStrings `json:"needBazelBuild,omitempty"`

	Run StringOrStrings `json:"run"`
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

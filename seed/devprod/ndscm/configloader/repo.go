package configloader

type RepoConfig struct {
	Upstream map[string]UpstreamConfig `json:"upstream,omitempty"`
}

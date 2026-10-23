package configloader

type RepoUpstreamConfig struct {
	Scm      string `json:"scm"`
	Repo     string `json:"repo,omitempty"`
	Local    bool   `json:"local,omitempty"`
	Tracking string `json:"tracking,omitempty"`
	Converge string `json:"converge"`
}

type RepoConfig struct {
	Upstream map[string]RepoUpstreamConfig `json:"upstream,omitempty"`
}

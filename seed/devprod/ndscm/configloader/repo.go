package configloader

type RepoNdscmConfig struct {
	Version string `json:"version,omitempty"`
}

type RepoUpstreamConfig struct {
	Scm      string `json:"scm"`
	Repo     string `json:"repo,omitempty"`
	Local    bool   `json:"local,omitempty"`
	Tracking string `json:"tracking,omitempty"`
	Converge string `json:"converge"`
}

type RepoConfig struct {
	Ndscm    RepoNdscmConfig               `json:"ndscm"`
	Upstream map[string]RepoUpstreamConfig `json:"upstream,omitempty"`
}

package configloader

type UpstreamRename struct {
	Match string `json:"match"`

	Replace string `json:"replace"`
}

type UpstreamConfig struct {
	Converge string `json:"converge"`

	Scm string `json:"scm"`

	Repo string `json:"repo,omitempty"`

	Local bool `json:"local,omitempty"`

	Tracking string `json:"tracking,omitempty"`

	Include []string `json:"include,omitempty"`

	Exclude []string `json:"exclude,omitempty"`

	Rename []UpstreamRename `json:"rename,omitempty"`
}

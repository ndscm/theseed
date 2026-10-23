package configloader

type UpstreamConfig struct {
	Converge string `json:"converge"`

	Scm string `json:"scm"`

	Repo string `json:"repo,omitempty"`

	Local bool `json:"local,omitempty"`

	Tracking string `json:"tracking,omitempty"`
}

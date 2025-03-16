package config

type Meta struct {
	Fails struct {
		List         string `json:"list"`
		Untrack      string `json:"untrack"`
		Cancel       string `json:"cancel"`
		Track        string `json:"track"`
		Registration string `json:"registration"`
		Unauthorized string `json:"unauthorized"`
		Unknown      string `json:"unknown"`
		Ack          string `json:"ack"`
	} `json:"fails"`

	Commands struct {
		Help    string `json:"help"`
		Track   string `json:"track"`
		Untrack string `json:"untrack"`
		List    string `json:"list"`
		Cancel  string `json:"cancel"`
		Start   string `json:"start"`
	} `json:"commands"`

	Descriptions []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"descriptions"`

	Replies struct {
		Registration string `json:"registration"`
		Track        string `json:"track"`
		Untracked    string `json:"untracked"`
		Untrack      string `json:"untrack"`
		Tags         string `json:"tags"`
		Filters      string `json:"filters"`
		Cancel       string `json:"cancel"`
		Tracking     string `json:"tracking"`
		TagsAck      string `json:"tags_ack"`
		FiltersAck   string `json:"filters_ack"`
	} `json:"replies"`

	Manuals struct {
		Link    string `json:"link"`
		Tags    string `json:"tags"`
		Filters string `json:"filters"`
		Acks    string `json:"acks"`
	} `json:"manuals"`

	Spans struct {
		Track   int `json:"track"`
		Untrack int `json:"untrack"`
		List    int `json:"list"`
	} `json:"spans"`

	Acks struct {
		Yes string `json:"yes"`
		No  string `json:"no"`
	} `json:"acknowledgment"`

	Services struct {
		GitHub                string `json:"github"`
		StackOverflow         string `json:"stackoverflow"`
		GitHubHost            string `json:"github_host"`
		GitHubBasePath        string `json:"github_base"`
		StackOverflowHost     string `json:"stackoverflow_host"`
		StackOverflowBasePath string `json:"stackoverflow_base"`
	}

	Updates struct {
		Plug string `json:"plug"`
	}
}

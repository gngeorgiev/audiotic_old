package player

type VlcStatus struct {
	Duration  int    `json:"length"`
	Time      int    `json:"time"`
	Volume    int    `json:"volume"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	State     string `json:"state"`
	Thumbnail string `json:"thumbnail"`
	IsPlaying bool   `json:"isPlaying"`
}

package api

import "gngeorgiev/audiotic/server/player"

func Status() (*player.VlcStatus, error) {
	return player.Get().Status()
}

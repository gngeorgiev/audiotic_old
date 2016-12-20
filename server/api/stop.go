package api

import "gngeorgiev/audiotic/server/player"

func Stop() error {
	return player.Get().Stop()
}

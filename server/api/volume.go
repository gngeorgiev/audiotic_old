package api

import "gngeorgiev/audiotic/server/player"

func Volume(v int) error {
	return player.Get().Volume(v)
}

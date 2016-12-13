package api

import "gngeorgiev/audiotic/server/player"

func Pause() error {
	return player.Get().Pause()
}

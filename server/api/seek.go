package api

import "gngeorgiev/audiotic/server/player"

func Seek(time int) error {
	return player.Get().Seek(time)
}

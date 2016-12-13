package api

import "gngeorgiev/audiotic/server/player"

func Resume() error {
	return player.Get().Resume()
}

package api

import "gngeorgiev/audiotic/server/player"

func Status() (*player.VlcStatusRoot, error) {
	s, err := player.Get().Status()
	if err != nil {
		return nil, err
	}

	return s.Root, err
}

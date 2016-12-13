package player

import vlc "github.com/adrg/libvlc-go"

var (
	player *vlc.Player
)

func InitPlayer() error {
	if player != nil {
		return nil
	}

	if err := vlc.Init("--no-video", "--quiet"); err != nil {
		return err
	}

	p, err := vlc.NewPlayer()
	if err != nil {
		return err
	}

	player = p
	return nil
}

func GetPlayer() *vlc.Player {
	return player
}

func RelasePlayer() {
	vlc.Release()
	player.Stop()
	player.Release()
}

func Stop() error {
	if player.IsPlaying() {
		return player.Stop()
	}

	return nil
}

func Play(url string) error {
	if err := Stop(); err != nil {
		return err
	}

	if err := player.SetMedia(url, false); err != nil {
		return err
	}

	if err := player.Play(); err != nil {
		return err
	}

	return nil
}

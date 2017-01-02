package player

import (
	"gngeorgiev/audiotic/server/models"

	"log"
	"time"

	vlc "github.com/adrg/libvlc-go"
)

func (v *VlcPlayer) createPlayer() (*vlc.Player, error) {
	player, err := vlc.NewPlayer()
	if err != nil {
		return nil, err
	}

	player.SetVolume(v.volume)
	return player, nil
}

func (v *VlcPlayer) updateStatus() {
	v.statsMutex.Lock()
	defer v.statsMutex.Unlock()

	if !v.mediaSet {
		return
	}

	st, err := v.player.MediaState()
	if err != nil {
		log.Println(err)
		return
	}
	v.state = st

	t, err := v.player.MediaTime()
	if err != nil {
		log.Println(err)
		return
	}
	v.time = t / 1000

	vol, err := v.player.Volume()
	if err != nil {
		log.Println(err)
		return
	}
	v.volume = vol

	v.isPlaying = v.player.IsPlaying() && v.state == vlc.MediaPlaying
}

func (v *VlcPlayer) update(t *time.Timer) {
	v.updateStatus()
	go v.notifyUpdated()
	t.Reset(updatesInterval)
}

func (v *VlcPlayer) play(t models.Track) error {
	if err := v.player.SetMedia(t.StreamUrl, false); err != nil {
		return err
	}

	if err := v.player.Play(); err != nil {
		return err
	}

	if err := v.waitForMediaState(vlc.MediaPlaying); err != nil {
		return err
	}

	d, err := v.player.MediaLength()
	if err != nil {
		return err
	}

	v.statsMutex.Lock()
	v.source = t.StreamUrl
	v.name = t.Title
	v.thumbnail = t.Thumbnail
	v.duration = d / 1000
	v.track = t
	v.mediaSet = true
	v.statsMutex.Unlock()
	return nil
}

func (v *VlcPlayer) pause() error {
	return v.player.SetPause(true)
}

func (v *VlcPlayer) resume() error {
	return v.player.SetPause(false)
}

func (v *VlcPlayer) seek(time int) error {
	return v.player.SetMediaTime(time * 1000)
}

func (v *VlcPlayer) stop() error {
	return v.player.Stop()
}

func (v *VlcPlayer) setVolume(vol int) error {
	return v.player.SetVolume(vol)
}

func (v *VlcPlayer) release() error {
	if err := v.Stop(); err != nil {
		return err
	}

	if v.player != nil {
		if err := v.player.Stop(); err != nil {
			return err
		}

		if err := v.player.Release(); err != nil {
			return err
		}
	}

	if err := vlc.Release(); err != nil {
		return err
	}

	return nil
}

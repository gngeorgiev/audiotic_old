package player

import (
	"gngeorgiev/audiotic/server/models"

	"log"
)

func (v *VlcPlayer) IsPlaying() bool {
	v.statsMutex.Lock()
	defer v.statsMutex.Unlock()

	return v.isPlaying
}

func (v *VlcPlayer) Resume() error {
	v.resumePlayingChan <- struct{}{}
	return nil
}

func (v *VlcPlayer) OnUpdated(ch chan *VlcStatus) {
	v.onUpdatedChansMutex.Lock()
	defer v.onUpdatedChansMutex.Unlock()

	v.onUpdatedChans = append(v.onUpdatedChans, ch)
}

func (v *VlcPlayer) Play(t models.Track) error {
	v.startedPlayingChan <- t
	return nil
}

func (v *VlcPlayer) Track() models.Track {
	v.statsMutex.Lock()
	defer v.statsMutex.Unlock()

	return v.track
}

func (v *VlcPlayer) Pause() error {
	v.pausedPlayingChan <- struct{}{}
	return nil
}

func (v *VlcPlayer) Stop() error {
	v.stoppedPlayingChah <- struct{}{}
	return nil
}

func (v *VlcPlayer) Seek(time int) error {
	v.seekChan <- time
	return nil
}

func (v *VlcPlayer) Volume(vol int) error {
	v.statsMutex.Lock()
	v.volume = vol
	v.statsMutex.Unlock()
	v.volumeChan <- vol
	return nil
}

func (v *VlcPlayer) Status() (*VlcStatus, error) {
	if v.player == nil {
		return nil, nil
	}

	status := &VlcStatus{}
	status.Name = v.name
	status.Duration = v.duration
	status.Source = v.source
	status.Time = v.time
	status.Volume = v.volume
	status.State = MediaStateToString(v.state)
	status.IsPlaying = v.IsPlaying()
	status.Thumbnail = v.thumbnail

	return status, nil
}

func (v *VlcPlayer) Release() error {
	v.releaseChan <- struct{}{}
	return nil
}

func Init() error {
	oncePlayer.Do(func() {
		player = &VlcPlayer{}
		if err := player.init(); err != nil {
			log.Fatal(err)
		}
	})

	return nil
}

func Get() *VlcPlayer {
	return player
}

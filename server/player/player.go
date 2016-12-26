package player

import (
	"sync"

	"time"

	"log"

	"fmt"

	"gngeorgiev/audiotic/server/models"

	"github.com/adrg/libvlc-go"
	"github.com/go-errors/errors"
)

type VlcHttpCommand string
type VlcHttpArgument struct {
	Name, Value string
}

type VlcPlayer struct {
	player *vlc.Player

	source, name, thumbnail string
	duration, time, volume  int
	state                   vlc.MediaState
	isPlaying               bool
	track                   models.Track
	statsMutex              sync.Mutex

	startedPlayingChan chan struct{}
	stoppedPlayingChah chan struct{}
	pausedPlayingChan  chan struct{}
	releaseChan        chan struct{}

	onUpdatedChansMutex sync.Mutex
	onUpdatedChans      []chan struct{}
}

var (
	oncePlayer sync.Once
	player     *VlcPlayer
)

func (v *VlcPlayer) init() error {
	if err := vlc.Init("--no-video", "--quiet"); err != nil {
		log.Fatal(err)
	}

	v.startedPlayingChan = make(chan struct{})
	v.stoppedPlayingChah = make(chan struct{})
	v.pausedPlayingChan = make(chan struct{})
	v.releaseChan = make(chan struct{})
	v.onUpdatedChans = make([]chan struct{}, 0)
	v.volume = 100

	go v.listenEvents()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			v.updateStatus()
		}
	}()

	return nil
}

func (v *VlcPlayer) createPlayer() (*vlc.Player, error) {
	player, err := vlc.NewPlayer()
	if err != nil {
		return nil, err
	}

	player.SetVolume(v.volume)
	return player, nil
}

func (v *VlcPlayer) IsPlaying() bool {
	return v.isPlaying
}

func (v *VlcPlayer) Resume() error {
	if v.player == nil {
		return nil
	}

	if !v.IsPlaying() {
		v.notifyStartPlaying()
		return v.player.SetPause(false)
	}

	return nil
}

func (v *VlcPlayer) updateStatus() {
	v.statsMutex.Lock()
	defer v.statsMutex.Unlock()
	if v.player == nil {
		return
	}

	t, err := v.player.MediaTime()
	if err != nil {
		log.Println(err)
		return
	}
	v.time = t / 1000

	st, err := v.player.MediaState()
	if err != nil {
		log.Println(err)
		return
	}
	v.state = st

	vol, err := v.player.Volume()
	if err != nil {
		log.Println(err)
		return
	}
	v.volume = vol

	v.isPlaying = v.player.IsPlaying() && v.state == vlc.MediaPlaying
}

func (v *VlcPlayer) update() {
	v.updateStatus()
	v.notifyUpdated()
}

func (v *VlcPlayer) listenEvents() {
	t := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-t.C:
			if v.player == nil {
				continue
			}

			v.update()
		case <-v.startedPlayingChan:
			v.update()
		case <-v.stoppedPlayingChah:
			v.update()
		case <-v.pausedPlayingChan:
			v.update()
		case <-v.releaseChan:
			t.Stop()
			return
		}
	}
}

func (v *VlcPlayer) notifyStartPlaying() {
	v.startedPlayingChan <- struct{}{}
}

func (v *VlcPlayer) notifyStopPlaying() {
	v.stoppedPlayingChah <- struct{}{}
}

func (v *VlcPlayer) notifyPausedPlaying() {
	v.pausedPlayingChan <- struct{}{}
}

func (v *VlcPlayer) notifyUpdated() {
	v.onUpdatedChansMutex.Lock()
	defer v.onUpdatedChansMutex.Unlock()

	payload := struct{}{}
	for _, ch := range v.onUpdatedChans {
		ch <- payload
	}
}

func (v *VlcPlayer) OnUpdated(ch chan struct{}) {
	v.onUpdatedChansMutex.Lock()
	defer v.onUpdatedChansMutex.Unlock()

	v.onUpdatedChans = append(v.onUpdatedChans, ch)
}

func (v *VlcPlayer) waitForMediaState(st vlc.MediaState) chan error {
	readyChan := make(chan error)
	go func() {
		t := time.NewTimer(60 * time.Second)

		for {
			select {
			case <-t.C:
				readyChan <- errors.New(fmt.Sprintf("Timeout waiting for state %s", MediaStateToString(st)))
				v.Stop()
				return
			default:
				if v.player == nil {
					continue
				}

				state, _ := v.player.MediaState()
				if state == st {
					t.Stop()
					readyChan <- nil
					return
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	return readyChan
}

func (v *VlcPlayer) Play(t models.Track) error {
	if v.player != nil {
		v.player.Stop()
		v.player.Release()
		v.player = nil
	}

	player, err := v.createPlayer()
	if err != nil {
		return err
	}
	v.player = player

	if err := v.player.SetMedia(t.StreamUrl, false); err != nil {
		return err
	}

	if err := v.player.Play(); err != nil {
		return err
	}

	if err := <-v.waitForMediaState(vlc.MediaPlaying); err != nil {
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
	v.statsMutex.Unlock()

	v.notifyStartPlaying()
	return nil
}

func (v *VlcPlayer) Track() models.Track {
	return v.track
}

func (v *VlcPlayer) Pause() error {
	if v.player == nil {
		return nil
	}

	if v.IsPlaying() {
		if err := v.player.SetPause(true); err != nil {
			return err
		}

		if err := <-v.waitForMediaState(vlc.MediaPaused); err != nil {
			return err
		}

		v.notifyPausedPlaying()
		return nil
	}

	return nil
}

func (v *VlcPlayer) Stop() error {
	if v.player == nil {
		return nil
	}

	if !v.IsPlaying() {
		return nil
	}

	if err := v.player.SetMediaTime(0); err != nil {
		return err
	}

	if err := v.Pause(); err != nil {
		return err
	}

	v.notifyStopPlaying()
	return nil
}

func (v *VlcPlayer) Seek(time int) error {
	if v.player == nil {
		return nil
	}

	return v.player.SetMediaTime(time * 1000)
}

func (v *VlcPlayer) Volume(vol int) error {
	v.statsMutex.Lock()
	v.volume = vol
	defer v.statsMutex.Unlock()

	return v.player.SetVolume(v.volume)
}

func (v *VlcPlayer) Release() error {
	defer func() {
		v.releaseChan <- struct{}{}
	}()

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

type VlcStatus struct {
	Duration  int    `json:"length"`
	Time      int    `json:"time"`
	Volume    int    `json:"volume"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	State     string `json:"state"`
	Thumbnail string `json:"thumbnail"`
	IsPlaying bool   `json:"isPlaying"`
}

func MediaStateToString(st vlc.MediaState) string {
	switch st {
	case vlc.MediaPlaying:
		return "playing"
	case vlc.MediaBuffering:
		return "buffering"
	case vlc.MediaEnded:
		return "ended"
	case vlc.MediaError:
		return "error"
	case vlc.MediaIdle:
		return "idle"
	case vlc.MediaOpening:
		return "openning"
	case vlc.MediaPaused:
		return "paused"
	case vlc.MediaStopped:
		return "stopped"
	default:
		return "unknown"
	}
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

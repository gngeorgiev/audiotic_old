package player

import (
	"sync"

	"time"

	"log"

	"fmt"

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
	duration, time          int
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

	player, err := vlc.NewPlayer()
	if err != nil {
		return err
	}

	v.player = player
	v.player.SetVolume(100)

	v.startedPlayingChan = make(chan struct{})
	v.stoppedPlayingChah = make(chan struct{})
	v.pausedPlayingChan = make(chan struct{})
	v.releaseChan = make(chan struct{})
	v.onUpdatedChans = make([]chan struct{}, 0)

	go v.listenEvents()
	return nil
}
func (v *VlcPlayer) IsPlaying() bool {
	return v.player.IsPlaying()
}

func (v *VlcPlayer) Resume() error {
	if !v.IsPlaying() {
		v.notifyStartPlaying()
		return v.player.SetPause(false)
	}

	return nil
}

func (v *VlcPlayer) listenEvents() {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-t.C:
			if v.IsPlaying() {
				v.notifyUpdated()
			}
		case <-v.startedPlayingChan:
			v.notifyUpdated()
		case <-v.stoppedPlayingChah:
			v.notifyUpdated()
		case <-v.pausedPlayingChan:
			v.notifyUpdated()
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
		t := time.NewTimer(30 * time.Second)

		for {
			select {
			case <-t.C:
				readyChan <- errors.New(fmt.Sprintf("Timeout waiting for state %s", mediaStateToString(st)))
				v.Stop()
				return
			default:
				state, err := v.player.MediaState()
				if err != nil {
					readyChan <- err
					return
				} else if state == st {
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

func (v *VlcPlayer) Play(source, name, thumbnail string) error {
	if v.IsPlaying() {
		v.Stop()
	}

	if err := v.player.SetMedia(source, false); err != nil {
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
	v.source = source
	v.name = name
	v.thumbnail = thumbnail
	v.duration = d / 1000
	v.statsMutex.Unlock()

	v.notifyStartPlaying()
	return nil
}

func (v *VlcPlayer) Pause() error {
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
	return v.player.SetMediaTime(time)
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
	}
	if err := vlc.Release(); err != nil {
		return err
	}

	return nil
}

type VlcStatus struct {
	Duration  int    `json:"length"`
	Time      int    `json:"time"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	State     string `json:"state"`
	Thumbnail string `json:"thumbnail"`
	IsPlaying bool   `json:"isPlaying"`
}

func mediaStateToString(st vlc.MediaState) string {
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
	status := &VlcStatus{}
	status.Name = v.name
	status.Duration = v.duration
	status.Source = v.source
	t, err := v.player.MediaTime()
	if err != nil {
		return nil, err
	}
	status.Time = t / 1000

	s, err := v.player.MediaState()
	if err != nil {
		return nil, err
	}

	status.IsPlaying = v.IsPlaying()
	status.State = mediaStateToString(s)
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

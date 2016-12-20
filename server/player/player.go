package player

import (
	"sync"

	"time"

	"strconv"

	"log"

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
	duration                int
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
			if !v.IsPlaying() {
				continue
			}

			v.notifyUpdated()
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

func (v *VlcPlayer) Play(source, name, thumbnail string, duration int) error {
	v.statsMutex.Lock()
	v.source = source
	v.name = name
	v.duration = duration
	v.thumbnail = thumbnail
	v.statsMutex.Unlock()

	if v.IsPlaying() {
		v.Stop()
	}

	if err := v.player.SetMedia(source, false); err != nil {
		return err
	}

	if err := v.player.Play(); err != nil {
		return err
	}

	readyChan := make(chan error)
	go func() {
		t := time.NewTimer(30 * time.Second)

		for {
			select {
			case <-t.C:
				readyChan <- errors.New("Timeout playing track")
				v.Stop()
				return
			default:
				state, err := v.player.MediaState()
				if err != nil {
					readyChan <- err
				}

				if state == vlc.MediaPlaying {
					v.notifyStartPlaying()
					t.Stop()
					readyChan <- nil
					return
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	return <-readyChan
}

func (v *VlcPlayer) Pause() error {
	if v.IsPlaying() {
		v.notifyPausedPlaying()
		return v.player.SetPause(true)
	}

	return nil
}

func (v *VlcPlayer) Stop() error {
	if v.IsPlaying() {
		v.notifyStopPlaying()
		return v.player.Stop()
	}

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
	Duration  string `json:"length"`
	Time      int    `json:"time"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	State     string `json:"state"`
	Thumbnail string `json:"thumbnail"`
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
	status.Duration = strconv.Itoa(v.duration)
	status.Source = v.source
	t, err := v.player.MediaTime()
	if err != nil {
		return nil, err
	}

	status.Time = t
	s, err := v.player.MediaState()
	if err != nil {
		return nil, err
	}

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

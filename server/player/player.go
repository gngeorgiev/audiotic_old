package player

import (
	"sync"

	"time"

	"log"

	"fmt"

	"gngeorgiev/audiotic/server/models"

	"runtime"

	"reflect"

	vlc "github.com/adrg/libvlc-go"
	"github.com/go-errors/errors"
)

type VlcPlayer struct {
	player *vlc.Player

	source, name, thumbnail string
	duration, time, volume  int
	state                   vlc.MediaState
	isPlaying, mediaSet     bool
	track                   models.Track
	statsMutex              sync.Mutex

	startedPlayingChan chan models.Track
	stoppedPlayingChah chan struct{}
	pausedPlayingChan  chan struct{}
	seekChan           chan int
	volumeChan         chan int
	resumePlayingChan  chan struct{}
	releaseChan        chan struct{}

	onUpdatedChansMutex sync.Mutex
	onUpdatedChans      []chan *VlcStatus
	statusOnLastUpdate  *VlcStatus
}

var (
	oncePlayer          sync.Once
	player              *VlcPlayer
	updatesInterval     = 500 * time.Millisecond
	maxWaitStateTimeout = 60 * time.Second
)

func (v *VlcPlayer) init() error {
	v.startedPlayingChan = make(chan models.Track)
	v.stoppedPlayingChah = make(chan struct{})
	v.pausedPlayingChan = make(chan struct{})
	v.resumePlayingChan = make(chan struct{})
	v.volumeChan = make(chan int)
	v.seekChan = make(chan int)
	v.releaseChan = make(chan struct{})
	v.onUpdatedChans = make([]chan *VlcStatus, 0)
	v.volume = 100

	go v.eventLoop()

	return nil
}

func (v *VlcPlayer) initInternals() error {
	if err := vlc.Init("--no-video"); err != nil {
		return err
	}

	player, err := v.createPlayer()
	if err != nil {
		return err
	}

	v.player = player
	return nil
}

func (v *VlcPlayer) eventLoop() {
	runtime.LockOSThread()
	if err := v.initInternals(); err != nil {
		log.Fatal(err)
	}

	t := time.NewTimer(updatesInterval)
	for {
		select {
		case <-t.C:
			v.update(t)
		case track := <-v.startedPlayingChan:
			if err := v.play(track); err != nil {
				log.Println(err)
			}

			v.update(t)
		case <-v.stoppedPlayingChah:
			if err := v.stop(); err != nil {
				log.Println(err)
			}

			v.update(t)
		case <-v.pausedPlayingChan:
			if err := v.pause(); err != nil {
				log.Println(err)
			}

			v.update(t)
		case <-v.resumePlayingChan:
			if err := v.resume(); err != nil {
				log.Println(err)
			}

			v.update(t)
		case vol := <-v.volumeChan:
			if err := v.setVolume(vol); err != nil {
				log.Println(err)
			}

			v.update(t)
		case pos := <-v.seekChan:
			if err := v.seek(pos); err != nil {
				log.Println(err)
			}

			v.update(t)
		case <-v.releaseChan:
			if err := v.release(); err != nil {
				log.Println(err)
			}

			return
		}
	}
}

func (v *VlcPlayer) notifyUpdated() {
	v.onUpdatedChansMutex.Lock()
	defer v.onUpdatedChansMutex.Unlock()

	payload, _ := v.Status()
	if !reflect.DeepEqual(v.statusOnLastUpdate, payload) {
		v.statusOnLastUpdate = payload
		for _, ch := range v.onUpdatedChans {
			ch <- payload
		}
	}
}

func (v *VlcPlayer) waitForMediaState(targetStates ...vlc.MediaState) error {
	timeoutTimer := time.NewTimer(maxWaitStateTimeout)
	updateTicker := time.NewTicker(updatesInterval)
	defer func() {
		timeoutTimer.Stop()
		updateTicker.Stop()
	}()

	for {
		select {
		case <-timeoutTimer.C:
			v.Stop()
			return errors.New(fmt.Sprintf("Timeout waiting for state %s", targetStates))
		case <-updateTicker.C:
			playerState, err := v.player.MediaState()
			if err != nil {
				log.Println(err)
				continue
			}

			for _, targetState := range targetStates {
				if targetState == playerState {
					return nil
				}
			}
		}
	}

	return nil
}

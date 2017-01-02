package api

import (
	"gngeorgiev/audiotic/server/player"
	"log"

	"sync"

	"github.com/adrg/libvlc-go"
)

var (
	autoplayEnabled  = make(chan struct{})
	autoplayDisabled = make(chan struct{})

	isAutoplayEnabled   bool
	autoplayInitialized bool
	mutex               sync.Mutex
)

func Autoplay(enabled bool) {
	mutex.Lock()
	defer mutex.Unlock()

	if !autoplayInitialized {
		go initAutoplay()
		autoplayInitialized = true
	}

	if isAutoplayEnabled == enabled {
		return
	}

	if enabled {
		autoplayEnabled <- struct{}{}
	} else {
		autoplayDisabled <- struct{}{}
	}
}

func initAutoplay() {
	for {
		<-autoplayEnabled
		p := player.Get()
		updatesCh := make(chan *player.VlcStatus)
		p.OnUpdated(updatesCh)

		for {
			select {
			case status := <-updatesCh:
				t := p.Track()
				if status.State == player.MediaStateToString(vlc.MediaEnded) {
					//if st.Time >= st.Duration && t.Provider != "" && t.Next != "" {
					if err := Play(t.Provider, t.Next); err != nil {
						log.Println(err)
					}
				}
			case <-autoplayDisabled:
				close(updatesCh)
				break
			}
		}
	}
}

package api

import (
	"gngeorgiev/audiotic/server/player"
	"log"
	"time"

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
		time.Sleep(500 * time.Millisecond)

		<-autoplayEnabled
		p := player.Get()
		updatesCh := make(chan struct{})
		p.OnUpdated(updatesCh)

		for {
			select {
			case <-updatesCh:
				st, _ := p.Status()
				t := p.Track()
				if st.State == player.MediaStateToString(vlc.MediaEnded) && t.Provider != "" && t.Next != "" {
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

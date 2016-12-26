package api

import (
	"fmt"
	"gngeorgiev/audiotic/server/player"
	"gngeorgiev/audiotic/server/providers"
	"strings"

	"gngeorgiev/audiotic/server/history"

	"time"

	"log"

	"github.com/go-errors/errors"
)

func Play(providerName, id string) error {
	p := providers.Container().GetComponent(func(p interface{}) bool {
		provider := p.(providers.Provider)
		return strings.ToLower(provider.GetName()) == strings.ToLower(providerName)
	})

	if p == nil {
		return errors.New(fmt.Sprintf("Unknown provider - %s", providerName))
	}

	provider := p.(providers.Provider)
	track, err := provider.Resolve(id)
	if err != nil {
		return err
	}

	if err := player.Get().Play(track); err != nil {
		log.Println(err)
	}

	track.StreamUrl = ""
	track.LastPlayed = time.Now()
	if err := history.Add(&track); err != nil {
		return err
	}

	return nil
}

package api

import (
	"fmt"
	"gngeorgiev/audiotic/server/player"
	"gngeorgiev/audiotic/server/providers"
	"strings"

	"gngeorgiev/audiotic/server/history"

	"github.com/go-errors/errors"
)

func Play(providerName, id string) error {
	p := providers.Container().GetComponent(func(p interface{}) bool {
		provider := p.(providers.Provider)
		return strings.ToLower(provider.GetName()) == providerName
	})

	if p == nil {
		return errors.New(fmt.Sprintf("Unknown provider - %s", providerName))
	}

	provider := p.(providers.Provider)
	track, err := provider.Resolve(id)
	if err != nil {
		return err
	}

	if err := player.Get().Play(track.StreamUrl, track.Title, track.Thumbnail); err != nil {
		return err
	}

	track.StreamUrl = ""
	if err := history.Add(&track); err != nil {
		return err
	}

	return nil
}

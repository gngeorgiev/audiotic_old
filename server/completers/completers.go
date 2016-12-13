package completers

import "gngeorgiev/audiotic/server/componentContainer"

type Completer interface {
	Complete(query string) ([]interface{}, error)
}

var (
	container = componentContainer.NewComponentContainer()
)

func Container() componentContainer.ComponentContainer {
	return container
}

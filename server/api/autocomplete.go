package api

import "gngeorgiev/audiotic/server/completers"

func Autocomplete(query string) ([]interface{}, error) {
	compl := completers.Container().GetComponents()
	return performConcurrentOperation(compl, func(c interface{}) ([]interface{}, error) {
		completer := c.(completers.Completer)
		return completer.Complete(query)
	})
}

package api

import "gngeorgiev/audiotic/server/providers"

func Search(query string) ([]interface{}, error) {
	prov := providers.Container().GetComponents()
	return performConcurrentOperation(prov, func(p interface{}) ([]interface{}, error) {
		provider := p.(providers.Provider)

		tracks, err := provider.Search(query)
		if err != nil {
			return nil, err
		}

		result := make([]interface{}, len(tracks))
		for i, t := range tracks {
			result[i] = t
		}

		return result, nil
	})
}

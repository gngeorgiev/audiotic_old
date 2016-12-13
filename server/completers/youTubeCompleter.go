package completers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	completeUrl = "http://suggestqueries.google.com/complete/search?client=firefox&ds=yt&q=%s"
)

func init() {
	Container().RegisterComponent(&youtubeCompleter{})
}

type youtubeCompleter struct {
}

func (y *youtubeCompleter) Complete(query string) ([]interface{}, error) {
	if query == "" {
		return make([]interface{}, 0), nil
	}

	autocompleteUrl := fmt.Sprintf(completeUrl, url.QueryEscape(query))
	resp, err := http.Get(autocompleteUrl)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var autocompleteData []interface{}
	jsonErr := json.Unmarshal(body, &autocompleteData)
	if jsonErr != nil {
		return nil, jsonErr
	}

	if len(autocompleteData) > 0 {
		return autocompleteData[1].([]interface{}), nil
	}

	return autocompleteData, nil
}

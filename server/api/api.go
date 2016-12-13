package api

import (
	"bytes"
	"sync"

	"github.com/go-errors/errors"
)

func merge(arr []interface{}) []interface{} {
	res := make([]interface{}, 0)
	for _, e := range arr {
		nestedResults := e.([]interface{})
		res = append(res, nestedResults...)
	}

	return res
}

func mergeErrors(arr []error) error {
	errs := bytes.Buffer{}
	for _, e := range arr {
		errs.WriteString(e.Error())
	}

	if errs.Len() > 0 {
		return errors.New(errs.String())
	}

	return nil
}

type operationFunc func(i interface{}) ([]interface{}, error)

func performConcurrentOperation(arr []interface{}, f operationFunc) ([]interface{}, error) {
	wg := sync.WaitGroup{}
	mut := sync.Mutex{}
	results := make([]interface{}, 0)
	errs := make([]error, 0)
	for _, c := range arr {
		wg.Add(1)
		go func(c interface{}) {
			defer wg.Done()
			data, err := f(c)
			mut.Lock()
			defer mut.Unlock()
			if err != nil {
				errs = append(errs, err)
			} else {
				results = append(results, data)
			}
		}(c)
	}

	wg.Wait()

	return results, mergeErrors(errs)
}

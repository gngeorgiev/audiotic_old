package componentContainer

import "sync"

type ComponentContainer interface {
	RegisterComponent(c interface{})
	GetComponents() []interface{}
	GetComponent(f func(c interface{}) bool) interface{}
}

type componentContainer struct {
	mu             sync.Mutex
	componentsList []interface{}
}

func NewComponentContainer() ComponentContainer {
	return &componentContainer{
		mu:             sync.Mutex{},
		componentsList: make([]interface{}, 0),
	}
}

func (cont *componentContainer) RegisterComponent(c interface{}) {
	cont.mu.Lock()
	defer cont.mu.Unlock()

	cont.componentsList = append(cont.componentsList, c)
}

func (cont *componentContainer) GetComponents() []interface{} {
	return cont.componentsList
}

func (cont *componentContainer) GetComponent(f func(interface{}) bool) interface{} {
	for _, c := range cont.GetComponents() {
		if f(c) {
			return c
		}
	}

	return nil
}

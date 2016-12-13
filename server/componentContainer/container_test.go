package componentContainer

import "testing"

func newContainer() *componentContainer {
	return NewComponentContainer().(*componentContainer)
}

func TestRegisterAndGetComponentsWorks(t *testing.T) {
	c := newContainer()

	val := "value"

	c.RegisterComponent(val)

	if c.GetComponents()[0] != val {
		t.Fatalf("Register component does not work %s", val)
	}
}

func TestGetComponentWorks(t *testing.T) {
	c := newContainer()
	val := "val"
	c.RegisterComponent(val)

	found := c.GetComponent(func(c interface{}) bool {
		if c.(string) == val {
			return true
		}

		return false
	})

	if found == nil {
		t.Fatal("Not found value in container")
	}
}

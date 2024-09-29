package async

import (
	"testing"
	"time"

	"github.com/jamesrom/order/atomicbit"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("Forward", func(t *testing.T) {
		t.Parallel()
		m := new(Map[string, string])
		m.Set("hello") <- "world"
		name := m.Get("hello")
		assert.Equal(t, "world", <-name)
	})
	t.Run("Backward", func(t *testing.T) {
		t.Parallel()
		m := new(Map[string, string])
		name := m.Get("hello")
		m.Set("hello") <- "world"
		assert.Equal(t, "world", <-name)
	})
	t.Run("Multiple Reads", func(t *testing.T) {
		t.Parallel()
		m := new(Map[string, string])
		m.Set("hello") <- "world"
		name1 := m.Get("hello")
		name2 := m.Get("hello")
		name3 := m.Get("hello")
		assert.Equal(t, "world", <-name1)
		assert.Equal(t, "world", <-name2)
		assert.Equal(t, "world", <-name3)

		name4 := m.Get("hello")
		assert.Equal(t, "world", <-name4)
	})

	// This test shows how the order of simultaneous writes is not guaranteed.
	t.Run("Simultaneous Writes", func(t *testing.T) {
		t.Parallel()
		m := new(Map[string, string])
		name := m.Get("hello")

		setter1 := m.Set("hello")
		setter2 := m.Set("hello")

		setter1 <- "world"
		setter2 <- "universe"

		assert.Contains(t, []string{"world", "universe"}, <-name)
	})
}

func TestGetElseSet(t *testing.T) {
	t.Parallel()

	t.Run("Happy", func(t *testing.T) {
		m := NewMap(map[string]string{"hello": "world"})
		name1 := m.GetElseSet("hello", func() string { return "world" })
		assert.Equal(t, "world", <-name1)
	})

	t.Run("Producer Invoked", func(t *testing.T) {
		m := new(Map[string, string])
		name1 := m.GetElseSet("hello", func() string { return "world" })
		assert.Equal(t, "world", <-name1)
		name2 := m.Get("hello")
		assert.Equal(t, "world", <-name2)
	})

	t.Run("Lots", func(t *testing.T) {
		checkBit := atomicbit.New(false)
		m := new(Map[string, string])

		for i := 0; i < 100; i++ {
			go func() {
				m.GetElseSet("hello", func() string {
					if checkBit.Get() { // fail if this function is called more than once
						panic("Producer called more than once")
					} else {
						t.Log("Producer called first time")
					}
					checkBit.Flip()
					time.Sleep(time.Second)
					return "world"
				})
			}()
		}
		assert.True(t, checkBit.Get())
	})
}

func TestDelete(t *testing.T) {
	m := NewMap(map[string]string{"hello": "world"})
	m.Delete("hello")
	name := m.Get("hello")
	select {
	case v := <-name:
		t.Errorf("Expected no value, got %s", v)
	default:
		// happy path
	}
}

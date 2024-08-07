package event

import (
	"fmt"
	"sync"
)

type EventHandler func(msg interface{})

type EventManager struct {
	Handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewEventManager() *EventManager {
	return &EventManager{
		Handlers: make(map[string][]EventHandler),
	}
}

func (em *EventManager) Register(event string, h EventHandler) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if event == "" {
		return fmt.Errorf("event can not be empty")
	}

	if em.containsHandler(event, h) {
		return fmt.Errorf("the same handler cant be added twice")
	}

	em.Handlers[event] = append(em.Handlers[event], h)
	return nil
}

func (em *EventManager) UnRegister(event string, h EventHandler) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	handlers, found := em.Handlers[event]
	if !found {
		return fmt.Errorf("could not find event: %s", event)
	}

	for i, handler := range handlers {
		if fmt.Sprintf("%p", handler) == fmt.Sprintf("%p", h) {
			em.Handlers[event] = append(handlers[:i], handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("could not find handler for event: %s", event)
}

func (em *EventManager) Emit(event string, msg interface{}, done chan struct{}) {
	go func() {
		em.mu.RLock()
		defer em.mu.RUnlock()

		handlers, found := em.Handlers[event]
		if !found {
			fmt.Printf("could not find event: %s\n", event)
			if done != nil {
				close(done)
			}
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(handlers))

		for _, handler := range handlers {
			go func(h EventHandler) {
				defer wg.Done()
				h(msg)
			}(handler)
		}

		wg.Wait()
		if done != nil {
			close(done)
		}
	}()
}

func (em *EventManager) containsHandler(event string, h EventHandler) bool {

	handlers, found := em.Handlers[event]
	if !found {
		return false
	}

	for _, handler := range handlers {
		if fmt.Sprintf("%p", handler) == fmt.Sprintf("%p", h) {
			return true
		}
	}

	return false
}

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
			break
		}
	}

	return nil
}

func (em *EventManager) Emit(event string, msg interface{}) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	handlers, found := em.Handlers[event]
	if !found {
		fmt.Printf("could not find event: %s\n", event)
		return
	}

	for _, handler := range handlers {
		go handler(msg)
	}
}

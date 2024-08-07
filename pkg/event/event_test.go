package event

import (
	"testing"
)

func TestNewEventManager(t *testing.T) {
	em := NewEventManager()

	if em == nil || em.Handlers == nil {
		t.Error("NewEventManager or Handlers should not return nil")
	}

	if len(em.Handlers) != 0 {
		t.Error("NewEventManager should initialize Handlers map with length 0")
	}
}

func TestRegister(t *testing.T) {
	var event = "test_event"

	em := NewEventManager()

	handler1 := func(msg interface{}) {
		// handle event
	}

	handler2 := func(msg interface{}) {
		// handle event
	}

	err := em.Register(event, handler1)
	if err != nil {
		t.Errorf("Register through and error: %v", err)
	}

	if len(em.Handlers[event]) != 1 {
		t.Error("Register should add the handler to the Handlers map")
	}

	// Registering the same handler for the same event should return an error
	err = em.Register(event, handler1)
	if err == nil {
		t.Errorf("Register through and error: %v", err)

	}

	if len(em.Handlers[event]) != 1 {
		t.Error("Register added the same handler to the event loop.")
	}

	// Add second handler
	err = em.Register(event, handler2)
	if err != nil {
		t.Errorf("could not Register second handler")
	}

	if len(em.Handlers[event]) != 2 {
		t.Errorf("Handlers did not have 2 events")
	}
}

func TestUnRegister(t *testing.T) {
	var event = "test_event"

	em := NewEventManager()

	handler := func(msg interface{}) {
		// handle event
	}

	err := em.Register(event, handler)
	if err != nil {
		t.Errorf("Register returned an error: %v", err)
	}

	if len(em.Handlers) != 1 {
		t.Errorf("Register does not have an event: %v", err)
	}

	err = em.UnRegister(event, handler)
	if err != nil {
		t.Errorf("UnRegister returned an error: %v", err)
	}

	if len(em.Handlers[event]) != 0 {
		t.Error("UnRegister should remove the handler from the Handlers map")
	}

	// Unregistering a non-existent handler for an event should return an error
	err = em.UnRegister(event, handler)
	if err == nil {
		t.Error("UnRegister should return an error when unregistering a non-existent handler for an event")
	}
}

func TestEmit(t *testing.T) {
	var event = "test_event"

	em := NewEventManager()

	handler1Called := false
	handler1 := func(msg interface{}) {
		handler1Called = true
	}

	handler2Called := false
	handler2 := func(msg interface{}) {
		handler2Called = true
	}

	em.Register(event, handler1)
	em.Register(event, handler2)

	if len(em.Handlers[event]) != 2 {
		t.Errorf("Handlers did not have 2 events")
	}

	done := make(chan struct{})

	em.Emit(event, nil, done)

	<-done

	if !handler1Called {
		t.Error("Emit should call all registered handlers for the event")
	}

	if !handler2Called {
		t.Error("Emit should call all registered handlers for the event")
	}

	// Emitting an event with no registered handlers should not cause any errors
	em.Emit("event2", nil, nil)
}

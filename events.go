package events

import (
	"fmt"
	"reflect"
	"sync"
)

// default
var (
	defaultEventManager = New()
	Watch               = defaultEventManager.Watch
	WatchNum            = defaultEventManager.WatchNum
	WatchOnce           = defaultEventManager.WatchOnce
	WatchAsync          = defaultEventManager.WatchAsync
	WatchNumAsync       = defaultEventManager.WatchNumAsync
	WatchOnceAsync      = defaultEventManager.WatchOnceAsync
	HasEvent            = defaultEventManager.HasEvent
	Events              = defaultEventManager.Events
	UnWatch             = defaultEventManager.UnWatch
	UnWatchEvent        = defaultEventManager.UnWatchEvent
	Clear               = defaultEventManager.Clear
	Trigger             = defaultEventManager.Trigger
	Wait                = defaultEventManager.Wait
)

type EventManager struct {
	eventsHandlers map[string][]*Handler
	sync.Mutex
	wg sync.WaitGroup
}

type Handler struct {
	fn    reflect.Value
	num   int
	async bool
}

func New() *EventManager {
	return &EventManager{
		make(map[string][]*Handler),
		sync.Mutex{},
		sync.WaitGroup{},
	}
}

func (e *EventManager) watchImpl(event string, fn interface{}, handler *Handler) error {
	e.Lock()
	defer e.Unlock()
	if !(reflect.TypeOf(fn).Kind() == reflect.Func) {
		return fmt.Errorf("not function:%s", reflect.TypeOf(fn).Kind())
	}
	e.eventsHandlers[event] = append(e.eventsHandlers[event], handler)
	return nil
}

func (e *EventManager) Watch(event string, fn interface{}) error {
	return e.watchImpl(event, fn, &Handler{reflect.ValueOf(fn), -1, false})
}

func (e *EventManager) WatchNum(event string, fn interface{}, num int) error {
	if num <= 0 {
		return fmt.Errorf("invalid num:%d", num)
	}
	return e.watchImpl(event, fn, &Handler{reflect.ValueOf(fn), num, false})
}

func (e *EventManager) WatchOnce(event string, fn interface{}) error {
	return e.WatchNum(event, fn, 1)
}

func (e *EventManager) WatchAsync(event string, fn interface{}) error {
	return e.watchImpl(event, fn, &Handler{reflect.ValueOf(fn), -1, true})
}

func (e *EventManager) WatchNumAsync(event string, fn interface{}, num int) error {
	return e.watchImpl(event, fn, &Handler{reflect.ValueOf(fn), num, true})
}

func (e *EventManager) WatchOnceAsync(event string, fn interface{}) error {
	return e.WatchNumAsync(event, fn, 1)
}

func (e *EventManager) HasEvent(event string) bool {
	e.Lock()
	defer e.Unlock()
	_, ok := e.eventsHandlers[event]
	return ok
}

func (e *EventManager) Events() []string {
	e.Lock()
	defer e.Unlock()
	events := make([]string, 0)
	for event := range e.eventsHandlers {
		events = append(events, event)
	}
	return events
}

func (e *EventManager) UnWatch(event string, fn interface{}) error {
	e.Lock()
	defer e.Unlock()
	if _, ok := e.eventsHandlers[event]; ok && len(e.eventsHandlers[event]) > 0 {
		e.remove(event, reflect.ValueOf(fn))
		return nil
	}
	return fmt.Errorf("event not exist:%s", event)
}

func (e *EventManager) UnWatchEvent(event string) error {
	e.Lock()
	defer e.Unlock()
	if _, ok := e.eventsHandlers[event]; !ok {
		return fmt.Errorf("event not exist:%s", event)
	}
	delete(e.eventsHandlers, event)
	return nil
}

func (e *EventManager) Clear() {
	e.Lock()
	defer e.Unlock()
	e.eventsHandlers = make(map[string][]*Handler)
}

func (e *EventManager) Trigger(event string, args ...interface{}) {
	e.Lock()
	defer e.Unlock()
	if _, ok := e.eventsHandlers[event]; ok {
		for _, handler := range e.eventsHandlers[event] {
			if handler.num > 0 {
				handler.num--
				if handler.num == 0 {
					e.remove(event, handler.fn)
				}
			}

			if !handler.async {
				e.triggerImpl(handler, event, args...)
			} else {
				e.wg.Add(1)
				go e.triggerAsyncImpl(handler, event, args...)
			}
		}
	}
}

func (e *EventManager) triggerImpl(handler *Handler, event string, args ...interface{}) {
	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}

	handler.fn.Call(params)
}

func (e *EventManager) triggerAsyncImpl(handler *Handler, event string, args ...interface{}) {
	defer e.wg.Done()
	e.triggerImpl(handler, event, args...)
}

func (e *EventManager) remove(event string, fn interface{}) {
	if _, ok := e.eventsHandlers[event]; !ok {
		return
	}

	for i, handler := range e.eventsHandlers[event] {
		if handler.fn == fn {
			l := len(e.eventsHandlers[event])
			if l == 1 {
				delete(e.eventsHandlers, event)
			} else {
				e.eventsHandlers[event] = append(e.eventsHandlers[event][:i], e.eventsHandlers[event][i+1:]...)
			}
			break
		}
	}
}

func (e *EventManager) Wait() {
	e.wg.Wait()
}

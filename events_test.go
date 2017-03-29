package events

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	mgr := New()
	fn := func() {}
	if mgr.Watch("foo", fn) != nil {
		t.Error()
	}
	if mgr.Watch("foo", "what") == nil {
		t.Error()
	}

	if mgr.UnWatch("foo", fn) != nil {
		t.Error()
	}
	if mgr.UnWatch("foo", fn) == nil {
		t.Error()
	}

	n := 0
	if mgr.Watch("foo", func(delta int) {
		n += delta
	}) != nil {
		t.Error()
	}
	if mgr.Watch("foo", func(delta int) {
		n += (delta + 1)
	}) != nil {
		t.Error()
	}
	mgr.Trigger("foo", 2)
	if n != 5 {
		t.Errorf("expected %d, actual %d", 5, n)
	}

}

func TestWatchNum(t *testing.T) {
	mgr := New()
	n := 0
	mgr.WatchNum("foo", func(delta int) {
		n += delta
	}, 3)
	mgr.Trigger("foo", 2)
	mgr.Trigger("foo", 2)
	if !mgr.HasEvent("foo") {
		t.Error()
	}
	mgr.Trigger("foo", 2)
	mgr.Trigger("foo", 2)
	mgr.Trigger("foo", 2)
	if n != 6 {
		t.Errorf("expected %d, actual %d", 6, n)
	}

	if mgr.HasEvent("foo") {
		t.Error()
	}
}

func TestWatchAsyc(t *testing.T) {
	mgr := New()
	ch := make(chan string)
	fn := func(str string, ch chan<- string) {
		ch <- str
	}
	mgr.WatchAsync("foo", fn)
	mgr.Trigger("foo", "a", ch)
	mgr.Trigger("foo", "b", ch)
	mgr.Trigger("foo", "c", ch)
	mgr.Trigger("foo", "d", ch)
	mgr.Trigger("foo", "e", ch)

	result := make([]string, 0)
	go func() {
		for str := range ch {
			result = append(result, str)
		}
	}()
	mgr.Wait()
	time.Sleep(2 * time.Millisecond)
	sort.Strings(result)
	if !reflect.DeepEqual(result, []string{"a", "b", "c", "d", "e"}) {
		t.Errorf("expected %s, actual %s", []string{"a", "b", "c", "d", "e"}, result)
	}
}

func TestDefalutEventManager(t *testing.T) {
	n := 0
	Watch("foo", func(delta int) {
		n += delta
	})
	Watch("foo", func(delta int) {
		n += (delta + 1)
	})
	Trigger("foo", 2)
	if n != 5 {
		t.Errorf("expected %d, actual %d", 5, n)
	}
}

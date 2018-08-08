// session_test.go
package session

import (
	"container/list"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	sm := createMgr(time.Minute)
	//1. one
	s := sm.Get("")
	if sm.list.Len() != 1 {
		t.Errorf("one: len != 1")
	}
	id1 := s.Id()
	// session GetSet..
	v := s.Get("key1")
	if v != nil {
		t.Errorf("one: v != nil")
	} else {
		s.Set("key1", "value1")
		if s.Get("key1") != "value1" {
			t.Errorf("one: should get 'value1")
		} else {
			s.Set("key1", "value2")
			if s.Get("key1") != "value2" {
				t.Errorf("one: should update to 'value2")
			} else {
				s.Remove("key1")
				if s.Get("key1") != nil {
					t.Errorf("one: should get nil after remove")
				}
			}
		}
	}

	//k1-v1, k2-v2 => k1-v3, k2-v4
	s.Set("k1", "v1")
	s.Set("k2", "v2")
	if s.Get("k1") != "v1" || s.Get("k2") != "v2" {
		t.Errorf("one: k1k2 fail")
	}
	s.Set("k1", "v3")
	s.Set("k2", "v4")
	if s.Get("k1") != "v3" || s.Get("k2") != "v4" {
		t.Errorf("one: k1k2 fail again")
	}

	//2. two: id2 id1
	s2 := sm.Get("")
	id2 := s2.Id()
	if sm.list.Len() != 2 {
		t.Errorf("two: len != 2")
	}
	if id2 != sm.list.Front().Value.(ses).id || id2 != s2.Id() {
		t.Errorf("two: session2 should be on head")
	}

	// id1
	s2.Invalidate()
	if sm.list.Len() != 1 {
		t.Errorf("two: invalidate not delete element")
	} else if id1 != sm.list.Front().Value.(ses).id {
		t.Errorf("two: invalidate left error")
	}

	//3. gc
	sm = createMgr(10 * time.Second) //10s timeout
	id1 = sm.Get("").Id()
	sm.gcOnce()
	if sm.list.Len() != 1 {
		t.Errorf("gc: should gc none")
	}
	// id2 id1(timed out)
	time.Sleep(12 * time.Second)
	id2 = sm.Get("").Id()
	sm.gcOnce()
	if sm.list.Len() != 1 || id2 != sm.list.Front().Value.(ses).id {
		t.Errorf("gc: should gc id1 and left id2")
	}
}

// 1w sessions with data: id9999 id9998 ... id0
func TestStress(t *testing.T) {
	sm := createMgr(10 * time.Second)
	var id0 string
	for i := 0; i < 10*1000; i++ {
		s := sm.Get("")
		if i == 0 {
			id0 = s.Id()
		}
		len := 1 + rand.Int()%10 // items
		for j := 0; j < len; j++ {
			len2 := 32 + rand.Int()%32
			s.Set(fmt.Sprintf("k%d", j), make([]byte, len2))
		}
	}
	if sm.list.Len() != 10*1000 {
		t.Errorf("stress: not 1w")
	}
	// get moves to front: id0 id9999 id9998 ...
	s0 := sm.Get(id0)
	if id0 != sm.list.Front().Value.(ses).id || id0 != s0.Id() {
		t.Errorf("stress: id0 not on head")
	}

	// id0 (id9999 id9998 ..)-all timed out
	time.Sleep(6 * time.Second)
	sm.Get(id0)
	time.Sleep(6 * time.Second)
	sm.gcOnce() // gcOnce has size limit (1000), so left more than id0
	if sm.list.Len() <= 1 || sm.list.Front().Value.(ses).Id() != id0 {
		t.Errorf("stress: all timed out fail. len=%d", sm.list.Len())
	}
}

// no GC started
func createMgr(d time.Duration) sesmgr {
	sm := sesmgr{
		lock:    new(sync.RWMutex),
		list:    new(list.List),
		smap:    make(map[string]*list.Element, 100),
		timeout: d,
	}
	return sm
}

// session.go
package session

import (
	"container/list"
	"crypto/rand"
	"fmt"
	"sync"
	"time"
	"net/http"
)

var TOKEN = "GSESSIONID" 						// session cookie name

type ISession interface {
	Id() string
	Get(key string) interface{}
	Set(key string, value interface{})
	Remove(key string)
	Invalidate()
}

type ISessionManager interface {
	Get(id string) ISession // get or create a session (with new id)
}

//#### #### ISession impl.

type ses struct {
	mgr  *sesmgr
	id   string
	lock *sync.RWMutex // for smap,time
	smap map[string]interface{}
	time time.Time //access time
}

func (s ses) Id() string {
	return s.id
}

func (s ses) Get(key string) interface{} {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.smap[key]
}

func (s ses) Set(key string, value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.smap[key] = value
}

func (s ses) Remove(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.smap, key)
}

func (s ses) Invalidate() {
	s.mgr.invalidate(s.id)
}

// If timed out, return false. Else update time and return true.
func (s ses) checkTimeout(d time.Duration) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if time.Now().Sub(s.time) > d {
		return false
	} else {
		s.time = time.Now()
		return true
	}
}

// Now - timeout - lastTime. >=0 means timed out.
func (s ses) getLeftTimeout(d time.Duration) time.Duration {
	s.lock.Lock()
	defer s.lock.Unlock()
	return time.Now().Sub(s.time) - d
}

//#### #### ISessionManager impl.

type sesmgr struct {
	lock    *sync.RWMutex
	list    *list.List               // a list of ses Element, active (top) -> inactive (bottom)
	smap    map[string]*list.Element // id => ses Element
	timeout time.Duration
}

func NewSessionManager(min int) ISessionManager {
	if min < 0 || min > 1000 {
		min = 30
	}
	sm := sesmgr{
		lock:    new(sync.RWMutex),
		list:    new(list.List),
		smap:    make(map[string]*list.Element, 100),
		timeout: time.Duration(min) * time.Minute,
	}
	go func() {
		for {
			time.Sleep(sm.gcOnce())
		}
	}()
	return sm
}

func (sm sesmgr) Get(id string) ISession {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if e, ok := sm.smap[id]; ok {
		s := e.Value.(ses)
		if s.checkTimeout(sm.timeout) {
			sm.list.MoveToFront(e) // front means most active
			return s
		} else {
			sm.delete(e)
		}
	}
	// not exists or timed out
	s := ses{
		mgr:  &sm,
		id:   genSesId(),
		lock: new(sync.RWMutex),
		smap: make(map[string]interface{}, 24),
		time: time.Now(),
	}
	e := sm.list.PushFront(s)
	sm.smap[s.id] = e
	return s
}

// (with no lock)
func (sm sesmgr) delete(e *list.Element) {
	v := sm.list.Remove(e)
	s := v.(ses)
	delete(sm.smap, s.id)
}

func (sm sesmgr) invalidate(id string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if e, ok := sm.smap[id]; ok {
		sm.delete(e)
	}
}

// generate a session id
func genSesId() string {
	ba := make([]byte, 32)
	if _, err := rand.Read(ba); err != nil {
		panic("session genId: " + err.Error())
	}
	return fmt.Sprintf("%x", ba)
}

//#### #### sesmgr GC

// Gc from bottom (least active) up. Return the pause before next round.
func (sm sesmgr) gcOnce() time.Duration {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	for i := 0; i < 1000; i++ { // max 1000 del
		e := sm.list.Back()
		if e == nil {
			break
		}
		s := e.Value.(ses)
		if d := s.getLeftTimeout(sm.timeout); d >= 0 {
			sm.delete(e)
		} else {
			if -d < 2*time.Minute { // still valid, wait a bit longer
				return 2 * time.Minute
			} else {
				return -d
			}
		}
	}
	if sm.list.Len() > 0 { // assume more to gc, catch up
		return 1 * time.Second
	} else {
		return 2 * time.Minute
	}
}

var sesmgrs = NewSessionManager(1) // session manager
func GetSession(w http.ResponseWriter, req *http.Request) ISession {
	var id = ""
	if c, err := req.Cookie(TOKEN); err == nil {
		id = c.Value
	}
	ses := sesmgrs.Get(id)
	if ses.Id() != id { //new session
		http.SetCookie(w, &http.Cookie{
			Name:  TOKEN,
			Value: ses.Id(),
		})
	}
	return ses
}

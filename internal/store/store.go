package store

import (
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var s *Store

func init() {
	internal_s := newStore()
	s = &internal_s
}

type Store struct {
	data map[string][]client.Object
	l    sync.RWMutex
}

func Get() *Store {
	return s
}

func newStore() Store {
	return Store{
		data: make(map[string][]client.Object, 0),
	}
}

func (s *Store) getKey(obj client.Object) string {
	return obj.GetNamespace() + "_" + obj.GetName()
}
func (s *Store) append(key string, u client.Object) {
	if uList, ok := s.data[key]; ok {
		uList = append(uList, u)
		s.data[key] = uList
	} else {
		s.data[key] = []client.Object{u}
	}
}

func (s *Store) Add(wList []client.Object, uList []client.Object) {
	for _, w := range wList {
		for _, u := range uList {
			s.l.Lock()
			s.append(s.getKey(w), u)
			s.l.Unlock()
		}
	}
}
func (s *Store) Delete(objList []client.Object) {
	for _, o := range objList {
		s.l.Lock()
		delete(s.data, s.getKey(o))
		s.l.Unlock()
	}
}
func (s *Store) Update(wList []client.Object, uList []client.Object) {
	s.Delete(wList)
	for _, w := range wList {
		for _, u := range uList {
			s.l.Lock()
			s.append(s.getKey(w), u)
			s.l.Unlock()
		}
	}
}

func (s *Store) IsInWatchList(obj client.Object) bool {
	s.l.RLock()
	defer s.l.RUnlock()
	uList, ok := s.data[s.getKey(obj)]
	return ok && len(uList) > 0
}
func (s *Store) GetUpdateList(watchObj client.Object) []client.Object {
	s.l.RLock()
	defer s.l.RUnlock()
	u, ok := s.data[s.getKey(watchObj)]
	if !ok {
		return nil
	}
	return u
}

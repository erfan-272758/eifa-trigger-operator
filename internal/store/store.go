package store

import (
	"slices"
	"strings"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var s *Store

func init() {
	internal_s := newStore()
	s = &internal_s
}

type Store struct {
	wToU   map[string][]client.Object
	wuToET map[string][]client.Object
	l      sync.RWMutex
}

func Get() *Store {
	return s
}

func newStore() Store {
	return Store{
		wToU:   make(map[string][]client.Object, 0),
		wuToET: make(map[string][]client.Object, 0),
	}
}

func (s *Store) getKey(objs ...client.Object) string {
	keys := make([]string, 0, len(objs))
	for _, obj := range objs {
		keys = append(keys, obj.GetNamespace()+"_"+obj.GetName())
	}
	return strings.Join(keys, ":")

}
func (s *Store) append(data map[string][]client.Object, key string, o client.Object) {
	if l, ok := data[key]; ok {
		// append if not exist
		if slices.IndexFunc(l, func(oldO client.Object) bool { return s.getKey(o) == s.getKey(oldO) }) == -1 {
			l = append(l, o)
			data[key] = l
		}

	} else {
		data[key] = []client.Object{o}
	}

}
func (s *Store) appendToUpdate(key string, u client.Object) {
	s.append(s.wToU, key, u)
}
func (s *Store) appendToET(key string, et client.Object) {
	s.append(s.wuToET, key, et)
}

func (s *Store) Delete(watchList []client.Object) {
	for _, w := range watchList {
		// get update lists
		uList := s.GetUpdateList(w)

		for _, u := range uList {
			// delete et from {watch,update}
			s.l.Lock()
			delete(s.wuToET, s.getKey(w, u))
			s.l.Unlock()
		}

		// delete update from watch
		s.l.Lock()
		delete(s.wToU, s.getKey(w))
		s.l.Unlock()
	}
}
func (s *Store) Update(et client.Object, wList []client.Object, uList []client.Object) {
	s.Delete(wList)
	for _, w := range wList {
		for _, u := range uList {
			s.l.Lock()
			// append to map watch=>update
			s.appendToUpdate(s.getKey(w), u)

			// append to map {watch,update}=>et
			s.appendToET(s.getKey(w, u), et)
			s.l.Unlock()

		}
	}
}

func (s *Store) IsInWatchList(w client.Object) bool {
	s.l.RLock()
	defer s.l.RUnlock()
	uList, ok := s.wToU[s.getKey(w)]
	return ok && len(uList) > 0
}
func (s *Store) GetUpdateList(watchObj client.Object) []client.Object {
	s.l.RLock()
	defer s.l.RUnlock()
	uList, ok := s.wToU[s.getKey(watchObj)]
	if !ok {
		return nil
	}
	return uList
}
func (s *Store) GetETList(watchObj client.Object, updateObj client.Object) []client.Object {
	s.l.RLock()
	defer s.l.RUnlock()
	etList, ok := s.wuToET[s.getKey(watchObj, updateObj)]
	if !ok {
		return nil
	}
	return etList
}

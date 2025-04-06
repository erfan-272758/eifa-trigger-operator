package store

import (
	triggerv1 "github.com/erfan-272758/eif-trigger-operator/api/v1"
)

var s *Store

func init() {
	internal_s := New()
	s = &internal_s
}

type Store struct {
}

func Get() *Store {
	return s
}

func New() Store {
	return Store{}
}

func (wr *Store) Add(et *triggerv1.EifaTrigger)    {}
func (wr *Store) Delete(et *triggerv1.EifaTrigger) {}
func (wr *Store) Update(et *triggerv1.EifaTrigger) {}
func (wr *Store) IsInWatchList() bool {
	return false
}

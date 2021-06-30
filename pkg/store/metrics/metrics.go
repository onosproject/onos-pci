// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package metrics

import (
	"context"
	"sync"

	e2smrcpre "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"

	ransim_types "github.com/onosproject/onos-api/go/onos/ransim/types"

	"github.com/onosproject/onos-lib-go/pkg/logging"

	"github.com/google/uuid"

	"github.com/onosproject/onos-lib-go/pkg/errors"
)

var log = logging.GetLogger("store", "metrics")

// Store kpm metrics store interface
type Store interface {
	Put(ctx context.Context, key uint64, entry Entry) (*Entry, error)

	// Get gets a metric store entry based on a given key
	Get(ctx context.Context, key uint64) (*Entry, error)

	// Update updates an existing entry in the store
	Update(ctx context.Context, key uint64, entry *Entry) error

	// UpdatePci only updates pci in the existing entry
	UpdatePci(ctx context.Context, key uint64, pci int32) error

	// Delete deletes an entry based on a given key
	Delete(ctx context.Context, key uint64) error

	// Entries list all of the metric store entries
	Entries(ctx context.Context, ch chan *Entry) error

	// Watch measurement store changes
	Watch(ctx context.Context, ch chan Event) error
}

type store struct {
	metrics  map[uint64]*Entry
	mu       sync.RWMutex
	watchers *Watchers
}

// NewStore creates new store
func NewStore() Store {
	watchers := NewWatchers()
	return &store{
		metrics:  make(map[uint64]*Entry),
		watchers: watchers,
	}
}

func (s *store) Entries(ctx context.Context, ch chan *Entry) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	defer close(ch)

	if len(s.metrics) == 0 {
		return errors.New(errors.NotFound, "no measurements entries stored")
	}

	for _, entry := range s.metrics {
		ch <- entry
	}
	return nil
}

func (s *store) Delete(ctx context.Context, key uint64) error {
	// TODO check the key and make sure it is not empty
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.metrics, key)
	return nil

}

func (s *store) Put(ctx context.Context, key uint64, entry Entry) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[key] = &entry
	s.watchers.Send(Event{
		Key:   key,
		Value: entry,
		Type:  Created,
	})
	return &entry, nil

}

func (s *store) Get(ctx context.Context, key uint64) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.metrics[key]; ok {
		return v, nil
	}
	return nil, errors.New(errors.NotFound, "the measurement entry does not exist")
}

func (s *store) Watch(ctx context.Context, ch chan Event) error {
	id := uuid.New()
	err := s.watchers.AddWatcher(id, ch)
	if err != nil {
		log.Error(err)
		close(ch)
		return err
	}
	go func() {
		<-ctx.Done()
		err = s.watchers.RemoveWatcher(id)
		if err != nil {
			log.Error(err)
		}
		close(ch)
	}()
	return nil
}

func (s *store) Update(ctx context.Context, key uint64, entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.metrics[key]; ok {
		s.metrics[key] = entry
		s.watchers.Send(Event{
			Key:   key,
			Value: *entry,
			Type:  Updated,
		})

		return nil
	}
	return errors.New(errors.NotFound, "the entry does not exist")
}

func (s *store) UpdatePci(ctx context.Context, key uint64, pci int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.metrics[key]; ok {
		s.metrics[key].Value.Metric.PCI = pci
		s.watchers.Send(Event{
			Key:   key,
			Value: *s.metrics[key],
			Type:  UpdatedPCI,
		})
		return nil
	}
	return errors.New(errors.NotFound, "the entry does not exist")
}

// NewKey creates a new measurements map key
func NewKey(cellGlobalID *e2smrcpre.CellGlobalId) uint64 {
	return nrcgiToInt(cellGlobalID.GetNrCgi())
}

// convert from NRCGI to uint64
func nrcgiToInt(nrcgi *e2smrcpre.Nrcgi) uint64 {
	array := nrcgi.PLmnIdentity.Value
	plmnid := uint32(array[0])<<0 | uint32(array[1])<<8 | uint32(array[2])<<16
	nci := nrcgi.NRcellIdentity.Value.Value
	return uint64(plmnid)<<36 | nci
}

func intToNRCGI(ncgi uint64) *e2smrcpre.Nrcgi {
	plmnid := uint32(ransim_types.GetPlmnID(ncgi))
	bitmask := uint32(0xFF)
	nci := uint64(ransim_types.GetNCI(ransim_types.NCGI(ncgi)))

	temp := &e2smrcpre.Nrcgi{
		PLmnIdentity: &e2smrcpre.PlmnIdentity{
			Value: []byte{byte((plmnid >> 0) & bitmask), byte((plmnid >> 8) & bitmask), byte((plmnid >> 16) & bitmask)},
		},
		NRcellIdentity: &e2smrcpre.NrcellIdentity{
			Value: &e2smrcpre.BitString{
				Value: nci,
				Len:   36,
			},
		},
	}
	return temp
}

var _ Store = &store{}

// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	e2smrcpre "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	"github.com/onosproject/onos-pci/pkg/types"
)

// Key metric key
type Key struct {
	CellGlobalID *e2smrcpre.CellGlobalId
}

// Entry entry of metrics store
type Entry struct {
	Key   Key
	Value types.CellPCI
}

// MetricEvent a metric event
type MetricEvent int

const (
	// None none cell event
	None MetricEvent = iota
	// Created created measurement event
	Created
	// Updated updated measurement event
	Updated
	// UpdatedPCI updated PCI in measurement
	UpdatedPCI
	// Deleted deleted measurement event
	Deleted
)

func (e MetricEvent) String() string {
	return [...]string{"None", "Created", "Updated", "UpdatedPCI", "Deleted"}[e]
}

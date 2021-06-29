// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package pci

import (
	"context"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-pci/pkg/manager"
	"github.com/onosproject/onos-pci/pkg/store/event"
	"github.com/onosproject/onos-pci/pkg/store/metrics"
	"github.com/onosproject/onos-pci/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (s *TestSuite) TestThreeCellPci(t *testing.T) {
	cfg := manager.Config{
		CAPath:      "/tmp/tls.cacrt",
		KeyPath:     "/tmp/tls.key",
		CertPath:    "/tmp/tls.crt",
		ConfigPath:  "/tmp/config.json",
		E2tEndpoint: "onos-e2t:5150",
		GRPCPort:    5150,
		SMName:      "oran-e2sm-rc-pre",
		SMVersion:   "v2",
	}

	_, err := certs.HandleCertPaths(cfg.CAPath, cfg.KeyPath, cfg.CertPath, true)
	assert.NoError(t, err)

	mgr := manager.NewManager(cfg)
	mgr.Run()

	// Get the metrics store and wait for an event indicating a change
	store := mgr.GetMetricsStore()
	ch := make(chan event.Event)
	err = store.Watch(context.Background(), ch)
	assert.NoError(t, err)

	// Anique PCI values.
	pcis := make(map[int32]int32)

	// Accrue unique PCI values. We start with two (one conflict in three cells) and will exit once we have
	// three unique values, which indicates that the PCI conflict was resolved.
	for e := range ch {
		pciEntry := e.Value.(*metrics.Entry).Value.(types.CellPCI)
		t.Log(fmt.Sprintf("Call %v has PCI %d", e.Key, pciEntry.Metric.PCI))
		pcis[pciEntry.Metric.PCI] = pciEntry.Metric.PCI
		if len(pcis) > 2 {
			t.Log("PCI conflict eliminated")
			break
		}
	}
}

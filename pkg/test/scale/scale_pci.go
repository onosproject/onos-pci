// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package scale

import (
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-pci/pkg/manager"
	"github.com/onosproject/onos-pci/pkg/store"
	"github.com/onosproject/onos-pci/pkg/test/utils"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func (s *TestSuite) TestScalePci(t *testing.T) {

	e2tEndpoint := "onos-e2t:5150"
	e2subEndpoint := "onos-e2sub:5150"

	cfg := manager.Config{
		CAPath:        "/tmp/tls.cacrt",
		KeyPath:       "/tmp/tls.key",
		CertPath:      "/tmp/tls.crt",
		ConfigPath:    "/tmp/config.json",
		E2tEndpoint:   e2tEndpoint,
		E2SubEndpoint: e2subEndpoint,
		GRPCPort:      5150,
		RicActionID:   int32(10),
		CtrlAcktimer:  5000,
	}

	_, err := certs.HandleCertPaths(cfg.CAPath, cfg.KeyPath, cfg.CertPath, true)
	assert.NoError(t, err)

	time.Sleep(10 * time.Second)
	resultCh := make(chan map[string]*store.PciStat)
	timer := make(chan bool)
	mgr := manager.NewManager(cfg)
	mgr.Run()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			resultCh <- mgr.Mons.PciMonitor
		}
	}()

	// timer
	go func() {
		time.Sleep(utils.TestTimeout)
		timer <- true
	}()

	for {
		numConflicts := int32(0)
		select {
		case <-timer:
			mgr.Mons.PciMonitorMutex.RLock()
			if mgr.Mons.PciMonitor["343332707642115"].NumConflicts >= 1 {
				numConflicts = mgr.Mons.PciMonitor["343332707642115"].NumConflicts
				mgr.Mons.PciMonitorMutex.RUnlock()
				break
			}
			if mgr.Mons.PciMonitor["343332707642118"].NumConflicts >= 1 {
				numConflicts = mgr.Mons.PciMonitor["343332707642118"].NumConflicts
				mgr.Mons.PciMonitorMutex.RUnlock()
				break
			}
			mgr.Mons.PciMonitorMutex.RUnlock()
			assert.NoError(t, fmt.Errorf("Timer experied"))
			os.Exit(1)

		case st := <-resultCh:
			mgr.Mons.PciMonitorMutex.RLock()
			if _, ok := st["343332707642115"]; !ok {
				continue
			}
			if _, ok := st["343332707642118"]; !ok {
				continue
			}
			if st["343332707642115"] == nil || st["343332707642118"] == nil {
				continue
			}

			log.Printf("num conflicts for %s is %d", "343332707642115", st["343332707642115"].NumConflicts)
			if st["343332707642115"].NumConflicts >= 1 {
				numConflicts = st["343332707642115"].NumConflicts
				mgr.Mons.PciMonitorMutex.RUnlock()
				break
			}
			log.Printf("num conflicts for %s is %d", "343332707642118", st["343332707642118"].NumConflicts)
			if st["343332707642118"].NumConflicts >= 1 {
				numConflicts = st["343332707642118"].NumConflicts
				mgr.Mons.PciMonitorMutex.RUnlock()
				break
			}
			mgr.Mons.PciMonitorMutex.RUnlock()

		}
		if numConflicts >= 1 {
			break
		}
	}
}

// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package scale

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-pci/pkg/manager"
	"log"
	"testing"
	"time"
)

func (s *TestSuite) TestScalePci(t *testing.T) {

	e2tEndpoint := "onos-e2t:5150"
	e2subEndpoint := "onos-e2sub:5150"

	cfg := manager.Config{
		CAPath:        "onos-pci/files/certs/tls.cacrt",
		KeyPath:       "onos-pci/files/certs/tls.key",
		CertPath:      "onos-pci/files/certs/tls.crt",
		ConfigPath:    "onos-pci/files/configs/config.json",
		E2tEndpoint:   e2tEndpoint,
		E2SubEndpoint: e2subEndpoint,
		GRPCPort:      5150,
		RicActionID:   int32(10),
	}

	_, err := certs.HandleCertPaths(cfg.CAPath, cfg.KeyPath, cfg.CertPath, true)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(10 * time.Second)
	//ready := make(chan bool)
	mgr := manager.NewManager(cfg)

	mgr.Run()
	for {
		time.Sleep(1 * time.Second)
		mgr.Mons.PciMonitorMutex.RLock()
		log.Printf("mgr.Mons.PciMonitor (length: %d): %v", len(mgr.Mons.PciMonitor), mgr.Mons.PciMonitor)
		for k, v := range mgr.Mons.PciMonitor {
			log.Printf("ID %s: %d\n", k, v)
		}
		mgr.Mons.PciMonitorMutex.RUnlock()

	}
	//<- ready
}

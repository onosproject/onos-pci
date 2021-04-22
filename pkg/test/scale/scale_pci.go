// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package scale

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-pci/pkg/manager"
	"testing"
	"time"
	"log"
)

func (s *TestSuite) TestScalePci(t *testing.T) {

	e2tEndpoint := "onos-e2t:5150"
	e2subEndpoint := "onos-e2sub:5150"

	cfg := manager.Config{
		CAPath:        "certs/tls.cacrt",
		KeyPath:        "certs/tls.key",
		CertPath:        "certs/tls.crt",
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
		log.Printf("mgr.Mons.PciMonitor: %v", mgr.Mons.PciMonitor)
		mgr.Mons.PciMonitorMutex.RUnlock()

	}
	//<- ready
}
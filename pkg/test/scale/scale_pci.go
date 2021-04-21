// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package scale

import (
	"github.com/onosproject/onos-pci/pkg/manager"
	"testing"
	"time"
)

func (s *TestSuite) TestScalePci(t *testing.T) {

	e2tEndpoint := "onos-e2t:5150"
	e2subEndpoint := "onos-e2sub:5150"

	cfg := manager.Config{
		E2tEndpoint:   e2tEndpoint,
		E2SubEndpoint: e2subEndpoint,
		GRPCPort:      5150,
		RicActionID:   int32(10),
	}

	time.Sleep(10 * time.Second)
	ready := make(chan bool)
	mgr := manager.NewManager(cfg)
	mgr.Run()
	<- ready
}
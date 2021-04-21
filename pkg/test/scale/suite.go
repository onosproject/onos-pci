// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package scale

import (
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-pci/pkg/test/utils"
)

// TestSuite is the primary onos-pci test suite
type TestSuite struct {
	test.Suite
}

// SetupTestSuite sets up the onos-pci test suite
func (s *TestSuite) SetupTestSuite() error {
	sdran, err := utils.CreateSdranRelease()
	if err != nil {
		return err
	}
	sdran.Set("ran-simulator.pci.metricName", "scale-rc-pci").
		Set("ran-simulator.pci.modelName", "scale-model")
	return sdran.Install(true)
}
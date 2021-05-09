// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package pci

import (
	"github.com/onosproject/helmit/pkg/helm"
	"github.com/onosproject/helmit/pkg/input"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-pci/pkg/test/utils"
)

// TestSuite is the primary onos-pci test suite
type TestSuite struct {
	sdran *helm.HelmRelease
	test.Suite
}

// SetupTestSuite sets up the onos-pci test suite
func (s *TestSuite) SetupTestSuite(c *input.Context) error {
	sdran, err := utils.CreateSdranRelease(c)
	s.sdran = sdran
	sdran.Set("ran-simulator.pci.metricName", "three-cell-metrics").
		Set("ran-simulator.pci.modelName", "three-cell-model")
	if err != nil {
		return err
	}
	return sdran.Install(true)
}

// TearDownTestSuite uninstalls helm chart released
func (s *TestSuite) TearDownTestSuite() error {
	return s.sdran.Uninstall()
}

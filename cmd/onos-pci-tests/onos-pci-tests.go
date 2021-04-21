// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"github.com/onosproject/helmit/pkg/registry"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-pci/pkg/test/pci"
	"github.com/onosproject/onos-pci/pkg/test/scale"
)

func main() {
	registry.RegisterTestSuite("pci", &pci.TestSuite{})
	registry.RegisterTestSuite("scale", &scale.TestSuite{})
	test.Main()
}
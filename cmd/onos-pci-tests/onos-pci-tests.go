// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/onosproject/helmit/pkg/registry"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-pci/test/pci"
	"github.com/onosproject/onos-pci/test/scale"
)

func main() {
	registry.RegisterTestSuite("pci", &pci.TestSuite{})
	registry.RegisterTestSuite("scale", &scale.TestSuite{})
	test.Main()
}

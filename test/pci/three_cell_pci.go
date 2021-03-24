// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package pci

import (
	e2tapi "github.com/onosproject/onos-api/go/onos/e2t/e2"
	"github.com/onosproject/onos-pci/pkg/controller"
	"github.com/onosproject/onos-pci/pkg/store"
	"github.com/onosproject/onos-ric-sdk-go/pkg/e2/indication"
	"github.com/stretchr/testify/assert"
	"github.com/onosproject/onos-pci/test/utils"
	"sync"
	"testing"
	"time"
)

func (s *TestSuite) TestThreeCellPci(t *testing.T) {
	sim := utils.CreateRanSimulatorWithNameOrDie(t, "e2epci")

	indCh := make(chan indication.Indication)

	e2IndCh := make(chan *store.E2NodeIndication)
	ctrlReqChMap := make(map[string]chan *e2tapi.ControlRequest)
	pciMonitor := make(map[string]*store.PciStat)
	pciMonMutex := &sync.RWMutex{}
	pciCtrl := controller.NewPciController(e2IndCh, ctrlReqChMap, pciMonitor, pciMonMutex)

	go pciCtrl.Run()

	sub, err := utils.CreatePciSubscriptionSingle(indCh, ctrlReqChMap)
	assert.NoError(t, err)

	var nodeIDs []string
	for k, _ := range ctrlReqChMap {
		nodeIDs = append(nodeIDs, k)
	}

	assert.Equal(t, 1, len(nodeIDs))

	numIndMsg := 0

	// Indication message block
	go func() {
		for {
			if numIndMsg >= 3 {
				t.Log("Received three indication messages - Succeed so far")
				break
			}
			select {
			case indMsg := <- indCh:
				e2IndCh <- &store.E2NodeIndication{
					NodeID: nodeIDs[0],
					IndMsg: indMsg,
				}
				numIndMsg++
			case <- time.After(10 * time.Second):
				t.Fatal("Indication message did not arrive before timer was expired")
			}
		}
	}()

	numCtrlMsg := 0

	// Control message block
	for {
		if numCtrlMsg >= 3 {
			t.Log("Received three control messages - Succeed so far")
			break
		}
		select {
		case ctrlMsg := <- ctrlReqChMap[nodeIDs[0]]:
			t.Logf("Received control message: %v", ctrlMsg)
			numCtrlMsg++
		case <- time.After(10 * time.Second):
			t.Fatal("Control message did not arrive before timer was expired")
		}
	}

	err = sub.Close()
	assert.NoError(t, err)

	err = sim.Uninstall()
	assert.NoError(t, err)
}

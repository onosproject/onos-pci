// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package manager

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	"github.com/onosproject/onos-ric-sdk-go/pkg/e2/indication"
)

var log = logging.GetLogger("manager")

// Config is a manager configuration
type Config struct {
	CAPath        string
	KeyPath       string
	CertPath      string
	E2tEndpoint   string
	E2SubEndpoint string
	GRPCPort      int
	AppConfig     *app.Config
	RicActionID   int32
}

// NewManager creates a new manager
func NewManager(config Config) *Manager {
	log.Info("Creating Manager")
	indCh := make(chan indication.Indication)
	return &Manager{
		Chans: Channels{
			IndCh: indCh,
		},
	}
}

// Manager is a manager for the KPIMON service
type Manager struct {
	Config   Config
	Sessions SBSessions
	Chans    Channels
	Ctrls    Controllers
	//	PeriodRange utils.PeriodRanges
}

// SBSessions is a set of Southbound sessions
type SBSessions struct{}

// Channels is a set of channels
type Channels struct {
	IndCh chan indication.Indication
}

// Controllers is a set of controllers
type Controllers struct{}

// Run starts the manager and the associated services
func (m *Manager) Run() {
	log.Info("Running Manager")
	if err := m.Start(); err != nil {
		log.Fatal("Unable to run Manager", err)
	}
}

// Start starts the manager
func (m *Manager) Start() error {
	return nil
}

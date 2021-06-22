// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package manager

import (
	"context"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/onos-pci/pkg/broker"
	appConfig "github.com/onosproject/onos-pci/pkg/config"
	"github.com/onosproject/onos-pci/pkg/controller"
	"github.com/onosproject/onos-pci/pkg/southbound/e2"
	"github.com/onosproject/onos-pci/pkg/store/metrics"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
)

var log = logging.GetLogger("manager")

// Config is a manager configuration
type Config struct {
	CAPath      string
	KeyPath     string
	CertPath    string
	ConfigPath  string
	E2tEndpoint string
	GRPCPort    int
	AppConfig   *app.Config
	SMName      string
	SMVersion   string
}

// NewManager creates a new manager
func NewManager(config Config) *Manager {
	appCfg, err := appConfig.NewConfig()
	if err != nil {
		log.Warn(err)
	}
	subscriptionBroker := broker.NewBroker()
	metricStore := metrics.NewStore()

	e2Manager, err := e2.NewManager(
		e2.WithE2TAddress("onos-e2t", 5150),
		e2.WithServiceModel(e2.ServiceModelName(config.SMName),
			e2.ServiceModelVersion(config.SMVersion)),
		e2.WithAppConfig(appCfg),
		e2.WithAppID("onos-pci"),
		e2.WithBroker(subscriptionBroker),
		e2.WithMetricStore(metricStore))

	if err != nil {
		log.Warn(err)
	}

	manager := &Manager{
		appConfig: appCfg,
		config:    config,
		e2Manager: e2Manager,
		pciCtrl:   controller.NewPciController(metricStore),
	}
	return manager
}

// Manager is a manager for the PCI xAPP service
type Manager struct {
	appConfig appConfig.Config
	config    Config
	e2Manager e2.Manager
	pciCtrl   controller.PciController
}

// Run starts the manager and the associated services
func (m *Manager) Run() {
	log.Info("Running Manager")
	if err := m.Start(); err != nil {
		log.Fatal("Unable to run Manager", err)
	}
}

// Start starts the manager
func (m *Manager) Start() error {
	// Start Northbound server
	err := m.startNorthboundServer()
	if err != nil {
		return err
	}

	err = m.e2Manager.Start()
	if err != nil {
		log.Warn(err)
		return err
	}

	m.pciCtrl.Run(context.Background())

	return nil
}

// Close kills the channels and manager related objects
func (m *Manager) Close() {
	log.Info("Closing Manager")
}

func (m *Manager) startNorthboundServer() error {
	s := northbound.NewServer(northbound.NewServerCfg(
		m.config.CAPath,
		m.config.KeyPath,
		m.config.CertPath,
		int16(m.config.GRPCPort),
		true,
		northbound.SecurityConfig{}))

	//s.AddService(nbi.NewService(m.Ctrls.PciCtrl))

	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		})
		if err != nil {
			doneCh <- err
		}
	}()
	return <-doneCh
}

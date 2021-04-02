# onos-pci
PCI xAPP for ONOS SD-RAN (µONOS Architecture)

## Overview 
The onos-pci is an xApp running over ONOS SD-RAN and as of now it supports the following features:

* Provides capability to subscribe to RC-PRE service model and receives indication messages from RAN simulator

* Provides capability to send control requests to change PCI values in RAN simulator

* Supports listing of PCI resources such metrics, neighbors, PCI, and PCI conflicts of cell(s) using CLI that is integrated with [onos-cli] 

* Detects PCI conflicts and resolves them based on an algorithm using cell neighbors information


See [README.md](docs/README.md) for details of running the onos-pci application.


[onos-cli]: https://github.com/onosproject/onos-cli
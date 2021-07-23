# Quick Start

## Prerequisite
Since ONOS SD-RAN has multiple micro-services running on the Kubernetes platform, 
onos-pci can run on the Kubernetes along with other ONOS SD-RAN micro-services. In order to deploy onos-kpimon, a Helm chart is necessary, which is in the 
[sdran-helm-charts] repository. 
Note that this application should be running together with the other SD-RAN micro-services such as Atomix, onos-operator, onos-e2t, onos-topo, onos-uenib, and onos-cli. sd-ran umbrella chart can be used
to deploy all essential micro-services and onos-pci.

The setup for PCI is similar to ransim, documented [here](https://github.com/onosproject/ran-simulator/blob/master/docs/quick_start.md). Instead of running 
```
helm install ran-simulator ran-simulator -n sd-ran
```
run
```
helm install onos-pci onos-pci -n sd-ran
```


[sdran-helm-charts]: https://github.com/onosproject/sdran-helm-charts
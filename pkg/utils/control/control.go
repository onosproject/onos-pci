// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/pdubuilder"
	e2smrccomm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
	e2smrc "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"github.com/onosproject/onos-lib-go/api/asn1/v1/asn1"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	ricStyleType             = 9
	controlActionID          = 1
	ranParamIDForPCI         = 1
	ranParamIDForNrPCI       = 11
	ranParamIDForEUTRAPCI    = 12
	ranParamIDForCGI         = 2
	ranParamIDForNrCGI       = 21
	ranParamIDForECGI        = 22
	ranParamIDForNrCGIPLMNID = 211
	ranParamIDForNrCGICellID = 212
	ranParamIDForECGIPLMNID  = 221
	ranParamIDForECGICellID  = 222
)

func CreateRcControlHeader(cgi *e2smrccomm.Cgi) ([]byte, error) {
	if cgi == nil {
		return nil, errors.NewInvalid("cgi is not set")
	}

	emptyUEID := &e2smrccomm.Ueid{
		Ueid: &e2smrccomm.Ueid_GNbUeid{
			GNbUeid: &e2smrccomm.UeidGnb{
				AmfUeNgapId: &e2smrccomm.AmfUeNgapId{Value: 0},
				Guami: &e2smrccomm.Guami{
					PLmnidentity: &e2smrccomm.Plmnidentity{Value: []byte{0, 0, 0}},
					AMfregionId:  &e2smrccomm.AmfregionId{Value: &asn1.BitString{Value: []byte{0}, Len: 8}},
					AMfsetId:     &e2smrccomm.AmfsetId{Value: &asn1.BitString{Value: []byte{0, 0}, Len: 10}},
					AMfpointer:   &e2smrccomm.Amfpointer{Value: &asn1.BitString{Value: []byte{0}, Len: 6}},
				},
				GNbCuUeF1ApIdList:   &e2smrccomm.UeidGnbCuF1ApIdList{Value: []*e2smrccomm.UeidGnbCuCpF1ApIdItem{&e2smrccomm.UeidGnbCuCpF1ApIdItem{GNbCuUeF1ApId: &e2smrccomm.GnbCuUeF1ApId{Value: 0}}}},
				GNbCuCpUeE1ApIdList: &e2smrccomm.UeidGnbCuCpE1ApIdList{Value: []*e2smrccomm.UeidGnbCuCpE1ApIdItem{&e2smrccomm.UeidGnbCuCpE1ApIdItem{GNbCuCpUeE1ApId: &e2smrccomm.GnbCuCpUeE1ApId{Value: 0}}}},
				RanUeid:             &e2smrccomm.Ranueid{Value: []byte{0, 0, 0, 0, 0, 0, 0, 0}},
				MNgRanUeXnApId:      &e2smrccomm.NgRannodeUexnApid{Value: 0},
				GlobalGnbId: &e2smrccomm.GlobalGnbId{
					PLmnidentity: &e2smrccomm.Plmnidentity{Value: []byte{0, 0, 0}},
					GNbId:        &e2smrccomm.GnbId{GnbId: &e2smrccomm.GnbId_GNbId{GNbId: &asn1.BitString{Value: []byte{0, 0, 0, 0}, Len: 32}}},
				},
				GlobalNgRannodeId: &e2smrccomm.GlobalNgrannodeId{
					GlobalNgrannodeId: &e2smrccomm.GlobalNgrannodeId_GNb{
						GNb: &e2smrccomm.GlobalGnbId{
							PLmnidentity: &e2smrccomm.Plmnidentity{Value: []byte{0, 0, 0}},
							GNbId:        &e2smrccomm.GnbId{GnbId: &e2smrccomm.GnbId_GNbId{GNbId: &asn1.BitString{Value: []byte{0, 0, 0, 0}, Len: 32}}},
						},
					},
				},
			},
		},
	}
	ctrlHdrFormat1, err := pdubuilder.CreateE2SmRcControlHeaderFormat1(emptyUEID, ricStyleType, controlActionID)
	if err != nil {
		return nil, err
	}

	err = ctrlHdrFormat1.Validate()
	if err != nil {
		return nil, err
	}

	protoBytes, err := proto.Marshal(ctrlHdrFormat1)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, nil
}

//E2SM-RC Control Message Format 1
//> List of RAN Parameters
//>> RAN Parameter ID: 1 (for Serving Cell PCI)
//>> Serving Cell PCI: RAN Parameter Value type: Structure
//>>> NR PCI: RAN Parameter Value type: Element with Key Flag False
//>>>> Integer (0...1007)
//>>> EUTRA PCI: RAN Parameter Value type: Element with Key Flag False
//>>>> Integer (0...503)
//>> RAN Parameter ID: 2 (for CGI)
//>> CGI: RAN Parameter Value Type: Structure
//>>> NR CGI: RAN Parameter Value type: Structure
//>>>> PLMN ID: RAN Parameter Value type: Element with Key Flag False
//>>>>> Octet string (size 3)
//>>>> NCI: RAN Parameter Value type: Element with Key Flag False
//>>>>> Bit string (size 36)
//>>>ECGI: RAN Parameter Value type: Structure
//>>>> PLMN ID: RAN Parameter Value type: Element with Key Flag False
//>>>>> Octet string (size 3)
//>>>> ECI: RAN Parameter Value type: Element with Key Flag False
//>>>>> Bit string (size 28)

func CreateRcControlMessage(pci int64, cgi *e2smrccomm.Cgi) ([]byte, error) {
	var scPciRanParamItem *e2smrc.E2SmRcControlMessageFormat1Item
	var cgiRanParamItem *e2smrc.E2SmRcControlMessageFormat1Item
	if cgi.GetNRCgi() != nil {
		// Serving Cell PCI
		nrPciRanParamValueInt, err := pdubuilder.CreateRanparameterValueInt(pci)
		if err != nil {
			return nil, err
		}
		nrPciRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(nrPciRanParamValueInt)
		if err != nil {
			return nil, err
		}
		nrPciRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForNrPCI, nrPciRanParamValue)
		if err != nil {
			return nil, err
		}
		scPciRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrc.RanparameterStructureItem{nrPciRanParamValueItem})
		if err != nil {
			return nil, err
		}
		scPciRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(scPciRanParamValue)
		if err != nil {
			return nil, err
		}
		scPciRanParamItem, err = pdubuilder.CreateE2SmRcControlMessageFormat1Item(ranParamIDForPCI, scPciRanParamValueType)
		if err != nil {
			return nil, err
		}
		// CGI
		nrCgiPlmnIDRANParamValueOcts, err := pdubuilder.CreateRanparameterValueOctS(cgi.GetNRCgi().GetPLmnidentity().GetValue())
		if err != nil {
			return nil, err
		}
		nrCgiPlmnIDRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(nrCgiPlmnIDRANParamValueOcts)
		if err != nil {
			return nil, err
		}
		nrCgiPlmnIDRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForNrCGIPLMNID, nrCgiPlmnIDRanParamValue)
		if err != nil {
			return nil, err
		}
		nrCgiCellIDRanParamValueBits, err := pdubuilder.CreateRanparameterValueBitS(cgi.GetNRCgi().GetNRcellIdentity().GetValue())
		if err != nil {
			return nil, err
		}
		nrCgiCellIDRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(nrCgiCellIDRanParamValueBits)
		if err != nil {
			return nil, err
		}
		nrCgiCellIDRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForNrCGICellID, nrCgiCellIDRanParamValue)
		if err != nil {
			return nil, err
		}
		nrCgiRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrc.RanparameterStructureItem{nrCgiPlmnIDRanParamValueItem, nrCgiCellIDRanParamValueItem})
		if err != nil {
			return nil, err
		}
		nrCgiRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(nrCgiRanParamValue)
		if err != nil {
			return nil, err
		}
		cgiRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForNrCGI, nrCgiRanParamValueType)
		if err != nil {
			return nil, err
		}
		cgiRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrc.RanparameterStructureItem{cgiRanParamValueItem})
		if err != nil {
			return nil, err
		}
		cgiRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(cgiRanParamValue)
		if err != nil {
			return nil, err
		}
		cgiRanParamItem, err = pdubuilder.CreateE2SmRcControlMessageFormat1Item(ranParamIDForCGI, cgiRanParamValueType)
		if err != nil {
			return nil, err
		}
	} else if cgi.GetEUtraCgi() != nil {
		// Serving Cell PCI
		ePciRanParamValueInt, err := pdubuilder.CreateRanparameterValueInt(pci)
		if err != nil {
			return nil, err
		}
		ePciRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(ePciRanParamValueInt)
		if err != nil {
			return nil, err
		}
		ePciRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForEUTRAPCI, ePciRanParamValue)
		if err != nil {
			return nil, err
		}
		scPciRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrc.RanparameterStructureItem{ePciRanParamValueItem})
		if err != nil {
			return nil, err
		}
		scPciRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(scPciRanParamValue)
		if err != nil {
			return nil, err
		}
		scPciRanParamItem, err = pdubuilder.CreateE2SmRcControlMessageFormat1Item(ranParamIDForPCI, scPciRanParamValueType)
		if err != nil {
			return nil, err
		}
		// CGI
		eCgiPlmnIDRANParamValueOcts, err := pdubuilder.CreateRanparameterValueOctS(cgi.GetEUtraCgi().GetPLmnidentity().GetValue())
		if err != nil {
			return nil, err
		}
		eCgiPlmnIDRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(eCgiPlmnIDRANParamValueOcts)
		if err != nil {
			return nil, err
		}
		eCgiPlmnIDRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForECGIPLMNID, eCgiPlmnIDRanParamValue)
		if err != nil {
			return nil, err
		}
		eCgiCellIDRanParamValueBits, err := pdubuilder.CreateRanparameterValueBitS(cgi.GetEUtraCgi().GetEUtracellIdentity().GetValue())
		if err != nil {
			return nil, err
		}
		eCgiCellIDRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(eCgiCellIDRanParamValueBits)
		if err != nil {
			return nil, err
		}
		eCgiCellIDRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForECGICellID, eCgiCellIDRanParamValue)
		if err != nil {
			return nil, err
		}
		eCgiRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrc.RanparameterStructureItem{eCgiPlmnIDRanParamValueItem, eCgiCellIDRanParamValueItem})
		if err != nil {
			return nil, err
		}
		eCgiRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(eCgiRanParamValue)
		if err != nil {
			return nil, err
		}
		cgiRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(ranParamIDForECGI, eCgiRanParamValueType)
		if err != nil {
			return nil, err
		}
		cgiRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrc.RanparameterStructureItem{cgiRanParamValueItem})
		if err != nil {
			return nil, err
		}
		cgiRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(cgiRanParamValue)
		if err != nil {
			return nil, err
		}
		cgiRanParamItem, err = pdubuilder.CreateE2SmRcControlMessageFormat1Item(ranParamIDForCGI, cgiRanParamValueType)
		if err != nil {
			return nil, err
		}
	}

	rpl := []*e2smrc.E2SmRcControlMessageFormat1Item{scPciRanParamItem, cgiRanParamItem}

	e2smRcControlMessage, err := pdubuilder.CreateE2SmRcControlMessageFormat1(rpl)
	if err != nil {
		return nil, err
	}

	err = e2smRcControlMessage.Validate()
	if err != nil {
		return nil, err
	}

	protoBytes, err := proto.Marshal(e2smRcControlMessage)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, nil
}

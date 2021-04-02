// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package utils

import (
	"context"
	modelapi "github.com/onosproject/onos-api/go/onos/ransim/model"
	"github.com/onosproject/onos-ric-sdk-go/pkg/e2/creds"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

const (
	ransimAddress = "ran-simulator:5150"
)

// AddDataSet loads a new data set and adds it to the specified data sets array
func AddDataSet(dataSets []*modelapi.DataSet, name string, path string) ([]*modelapi.DataSet, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dataSets = append(dataSets, &modelapi.DataSet{Type: name, Data: data})
	return dataSets, nil
}

// LoadNewModel loads the specified data sets into the RAN simulator
func LoadNewModel(dataSets []*modelapi.DataSet) error {
	client, conn, err := NewRansimClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = client.Load(context.Background(), &modelapi.LoadRequest{DataSet: dataSets, Resume: true})
	return err
}

// NewRansimClient returns a client for engaging with the RAN simulator API
func NewRansimClient() (modelapi.ModelServiceClient, *grpc.ClientConn, error) {
	tlsConfig, err := creds.GetClientCredentials()
	if err != nil {
		return nil, nil, err
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	conn, err := grpc.DialContext(context.Background(), ransimAddress, opts...)
	if err != nil {
		return nil, nil, err
	}
	return modelapi.NewModelServiceClient(conn), conn, nil
}

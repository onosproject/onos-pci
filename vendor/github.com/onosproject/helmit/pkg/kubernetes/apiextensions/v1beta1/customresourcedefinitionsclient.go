// Code generated by helmit-generate. DO NOT EDIT.

package v1beta1

import (
	"github.com/onosproject/helmit/pkg/kubernetes/resource"
)

type CustomResourceDefinitionsClient interface {
	CustomResourceDefinitions() CustomResourceDefinitionsReader
}

func NewCustomResourceDefinitionsClient(resources resource.Client, filter resource.Filter) CustomResourceDefinitionsClient {
	return &customResourceDefinitionsClient{
		Client: resources,
		filter: filter,
	}
}

type customResourceDefinitionsClient struct {
	resource.Client
	filter resource.Filter
}

func (c *customResourceDefinitionsClient) CustomResourceDefinitions() CustomResourceDefinitionsReader {
	return NewCustomResourceDefinitionsReader(c.Client, c.filter)
}

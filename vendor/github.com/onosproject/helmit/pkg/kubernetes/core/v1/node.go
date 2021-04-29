// Code generated by helmit-generate. DO NOT EDIT.

package v1

import (
	"github.com/onosproject/helmit/pkg/kubernetes/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"time"
)

var NodeKind = resource.Kind{
	Group:   "",
	Version: "v1",
	Kind:    "Node",
	Scoped:  false,
}

var NodeResource = resource.Type{
	Kind: NodeKind,
	Name: "nodes",
}

func NewNode(node *corev1.Node, client resource.Client) *Node {
	return &Node{
		Resource: resource.NewResource(node.ObjectMeta, NodeKind, client),
		Object:   node,
	}
}

type Node struct {
	*resource.Resource
	Object *corev1.Node
}

func (r *Node) Delete() error {
	client, err := kubernetes.NewForConfig(r.Config())
	if err != nil {
		return err
	}
	return client.CoreV1().
		RESTClient().
		Delete().
		NamespaceIfScoped(r.Namespace, NodeKind.Scoped).
		Resource(NodeResource.Name).
		Name(r.Name).
		VersionedParams(&metav1.DeleteOptions{}, metav1.ParameterCodec).
		Timeout(time.Minute).
		Do().
		Error()
}
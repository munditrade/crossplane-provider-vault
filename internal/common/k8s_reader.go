package common

import "sigs.k8s.io/controller-runtime/pkg/client"

type K8sReader interface {
	client.Reader
}

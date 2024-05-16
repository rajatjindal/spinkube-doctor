package minikube

import (
	"context"

	"github.com/rajatjindal/spinkube/pkg/checks"
	"github.com/rajatjindal/spinkube/pkg/provider"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type minikube struct {
	k8sclient kubernetes.Interface
	dc        dynamic.Interface
}

func New(dc dynamic.Interface, sc kubernetes.Interface) provider.Provider {
	return &minikube{
		k8sclient: sc,
		dc:        dc,
	}
}

func (k *minikube) Name() string {
	return "minikube"
}

func (k *minikube) Client() kubernetes.Interface {
	return k.k8sclient
}

func (k *minikube) DynamicClient() dynamic.Interface {
	return k.dc
}

func (k *minikube) Status(ctx context.Context) ([]provider.Status, error) {
	return checks.Status(ctx, k)
}

func (k *minikube) GetCheckOverride(ctx context.Context, check provider.Check) provider.CheckFn {
	switch check.Name {
	case checks.CheckBinaryInstalledOnNodes:
		return binaryVersionCheck
	}

	return nil
}

var binaryVersionCheck = func(ctx context.Context, k provider.Provider, check provider.Check) (provider.Status, error) {
	return checks.ExecOnEachNodeFn(ctx, k, check, []string{"/host/opt/kwasm/bin/containerd-shim-spin-v2", "-v"})
}

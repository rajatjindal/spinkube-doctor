package k3d

import (
	"context"

	"github.com/rajatjindal/spinkube/pkg/provider"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type k3d struct {
	k8sclient kubernetes.Interface
	dc        dynamic.Interface
}

func New(dc dynamic.Interface, sc kubernetes.Interface) provider.Provider {
	return &k3d{
		k8sclient: sc,
		dc:        dc,
	}
}

func (k *k3d) Name() string {
	return "k3d"
}

func (k *k3d) Status(ctx context.Context) ([]provider.Status, error) {
	statusList := []provider.Status{}

	for _, check := range []Check{
		containerdCheck,
		spinappCrdCheck,
		spinappExecutorCrdCheck,
		certManagerCrdCheck,
		runtimeClassCheck,
	} {
		status, err := check(ctx, k)
		if err != nil {
			return nil, err
		}
		statusList = append(statusList, status)
	}

	return statusList, nil
}

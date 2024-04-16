package k3d

import (
	"context"
	"fmt"

	"github.com/rajatjindal/spinkube/pkg/provider"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var checksMap = map[string]CheckFn{
	"crd":                         isCrdInstalled,
	"containerd-version-on-nodes": containerdVersionCheck,
	"runtimeclass":                runtimeClassCheck,
}

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

	checks := []Check{}
	err := yaml.Unmarshal(rawChecks, &checks)
	if err != nil {
		return nil, err
	}

	for _, check := range checks {
		// fmt.Printf("Running check %q\n", check.Name)

		checkfn, ok := checksMap[check.Type]
		if !ok {
			return nil, fmt.Errorf("check type %q not supported", check.Type)
		}

		status, err := checkfn(ctx, k, check)
		if err != nil {
			return nil, err
		}

		statusList = append(statusList, status)
	}

	return statusList, nil
}

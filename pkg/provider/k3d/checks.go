package k3d

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/rajatjindal/spinkube/pkg/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:embed checks/checks.yaml
var rawChecks []byte

type Check struct {
	Name         string   `yaml:"name"`
	Type         string   `yaml:"checkType"`
	ResourceName string   `yaml:"resourceName"`
	SemVer       []string `yaml:"semver"`
}

type CheckFn func(ctx context.Context, k *k3d, check Check) (provider.Status, error)

var isCrdInstalled = func(ctx context.Context, k *k3d, check Check) (provider.Status, error) {
	_, err := k.dc.Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}).Get(ctx, check.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return provider.Status{
				Name: check.Name,
				Ok:   false,
			}, nil
		}

		return provider.Status{
			Name: check.Name,
			Ok:   false,
		}, err
	}

	return provider.Status{
		Name: check.Name,
		Ok:   true,
	}, nil
}

var runtimeClassCheck = func(ctx context.Context, k *k3d, check Check) (provider.Status, error) {
	_, err := k.k8sclient.NodeV1().RuntimeClasses().Get(ctx, check.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return provider.Status{
				Name: check.Name,
				Ok:   false,
			}, nil
		}

		return provider.Status{
			Name: check.Name,
			Ok:   false,
		}, err
	}

	return provider.Status{
		Name: check.Name,
		Ok:   true,
	}, nil
}

var containerdVersionCheck = func(ctx context.Context, k *k3d, check Check) (provider.Status, error) {
	resp, err := k.k8sclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return provider.Status{}, err
	}

	vok := true
	msgs := []string{}

	for _, node := range resp.Items {
		version := strings.ReplaceAll(node.Status.NodeInfo.ContainerRuntimeVersion, "containerd://", "")
		ok, err := compareVersions(version, check.SemVer)
		if err != nil {
			vok = false
			msgs = append(msgs, err.Error())
			continue
		}

		if !ok {
			vok = false
			msgs = append(msgs, fmt.Sprintf("  - node: %s with containerd version %s does not support SpinApps", node.Name, node.Status.NodeInfo.ContainerRuntimeVersion))
			continue
		}
	}

	return provider.Status{
		Name: check.Name,
		Ok:   vok,
		Msg:  strings.Join(msgs, "\n"),
	}, nil
}

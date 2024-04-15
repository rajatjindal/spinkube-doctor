package k3d

import (
	"context"
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/rajatjindal/spinkube/pkg/provider"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	min1dot6Check, _ = semver.NewConstraint("~1.6.8-0")
	min1dot7Check, _ = semver.NewConstraint("~1.7.6-0")
)

var minVersionsCheck = []*semver.Constraints{
	min1dot6Check,
	min1dot7Check,
}

type Check func(ctx context.Context, k *k3d) (provider.Status, error)

var spinappCrdCheck = func(ctx context.Context, k *k3d) (provider.Status, error) {
	installed, err := isCrdInstalled(ctx, k.dc, "spinapps.core.spinoperator.dev")
	if err != nil {
		return provider.Status{}, err
	}

	return provider.Status{
		Name:      "SpinApp CRD",
		Ok:        installed,
		Installed: installed,
	}, nil
}

var spinappExecutorCrdCheck = func(ctx context.Context, k *k3d) (provider.Status, error) {
	installed, err := isCrdInstalled(ctx, k.dc, "spinappexecutors.core.spinoperator.dev")
	if err != nil {
		return provider.Status{}, err
	}

	return provider.Status{
		Name:      "SpinAppExecutor CRD",
		Ok:        installed,
		Installed: installed,
	}, nil
}

var certManagerCrdCheck = func(ctx context.Context, k *k3d) (provider.Status, error) {
	installed, err := isCrdInstalled(ctx, k.dc, "certificates.cert-manager.io")
	if err != nil {
		return provider.Status{}, err
	}

	return provider.Status{
		Name:      "CertManager Certificate CRD",
		Ok:        installed,
		Installed: installed,
	}, nil
}

var runtimeClassCheck = func(ctx context.Context, k *k3d) (provider.Status, error) {
	_, err := k.k8sclient.NodeV1().RuntimeClasses().Get(ctx, "wasmtime-spin-v2", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return provider.Status{
				Name:      "SpinApp Runtime Class",
				Ok:        false,
				Installed: false,
			}, nil
		}

		return provider.Status{}, err
	}

	return provider.Status{
		Name:      "SpinApp Runtime Class",
		Ok:        true,
		Installed: true,
	}, nil
}

var containerdCheck = func(ctx context.Context, k *k3d) (provider.Status, error) {
	resp, err := k.k8sclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return provider.Status{}, err
	}

	vok := true
	msgs := []string{}

	for _, node := range resp.Items {
		ok, msg, err := checkContainerdVersionOnNode(ctx, node)
		if err != nil {
			vok = false
		}

		if !ok {
			vok = false
			msgs = append(msgs, msg)
		}

	}

	return provider.Status{
		Name:      "Containerd Version",
		Ok:        vok,
		Installed: true,
		Msg:       strings.Join(msgs, "\n"),
	}, nil
}

func checkContainerdVersionOnNode(_ context.Context, node v1.Node) (bool, string, error) {
	version := strings.ReplaceAll(node.Status.NodeInfo.ContainerRuntimeVersion, "containerd://", "")
	actualVersion, err := semver.NewVersion(version)
	if err != nil {
		return false, "", err
	}

	for _, vcheck := range minVersionsCheck {
		k, err := vcheck.Validate(actualVersion)
		if len(err) > 0 {
			continue
		}

		return k, "", nil
	}

	return false, fmt.Sprintf("  - node: %s with containerd version %s does not support SpinApps", node.Name, version), nil
}

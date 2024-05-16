package checks

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rajatjindal/spinkube/pkg/provider"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var checksMap = map[string]provider.CheckFn{
	"crd":                         isCrdInstalled,
	"containerd-version-on-nodes": containerdVersionCheck,
	"runtimeclass":                runtimeClassCheck,
	"deployment-running":          deploymentRunningCheck,
	"binary-installed-on-nodes":   binaryVersionCheck,
}

//go:embed data/checks.yaml
var rawChecks []byte

var isCrdInstalled = func(ctx context.Context, k provider.Provider, check provider.Check) (provider.Status, error) {
	_, err := k.DynamicClient().Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}).Get(ctx, check.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return provider.Status{
				Name:     check.Name,
				Ok:       false,
				HowToFix: check.HowToFix,
			}, nil
		}

		return provider.Status{
			Name:     check.Name,
			Ok:       false,
			HowToFix: check.HowToFix,
		}, err
	}

	return provider.Status{
		Name: check.Name,
		Ok:   true,
	}, nil
}

var runtimeClassCheck = func(ctx context.Context, k provider.Provider, check provider.Check) (provider.Status, error) {
	_, err := k.Client().NodeV1().RuntimeClasses().Get(ctx, check.ResourceName, metav1.GetOptions{})
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

var deploymentRunningCheck = func(ctx context.Context, k provider.Provider, check provider.Check) (provider.Status, error) {
	resp, err := k.Client().AppsV1().Deployments(v1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return provider.Status{
				Name:     check.Name,
				Ok:       false,
				HowToFix: check.HowToFix,
			}, nil
		}

		return provider.Status{
			Name:     check.Name,
			Ok:       false,
			HowToFix: check.HowToFix,
		}, err
	}

	//TODO: handle pagination
	for _, item := range resp.Items {
		if item.Name == check.ResourceName {
			if len(check.SemVer) > 0 {
				imageTag := getImageTag(item, check)
				ok, err := compareVersions(imageTag, check.SemVer)
				if err != nil {
					return provider.Status{
						Name:     check.Name,
						Ok:       false,
						Msg:      fmt.Sprintf("deployment running, but failed to do version check: %v", err),
						HowToFix: check.HowToFix,
					}, nil
				}

				if !ok {
					return provider.Status{
						Name:     check.Name,
						Ok:       false,
						Msg:      fmt.Sprintf("deployment running, but version check failed: %v", err),
						HowToFix: check.HowToFix,
					}, nil
				}
			}

			return provider.Status{
				Name: check.Name,
				Ok:   true,
			}, nil
		}
	}

	return provider.Status{
		Name:     check.Name,
		Ok:       false,
		HowToFix: check.HowToFix,
	}, nil
}

var containerdVersionCheck = func(ctx context.Context, k provider.Provider, check provider.Check) (provider.Status, error) {
	resp, err := k.Client().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return provider.Status{}, err
	}

	vok := true
	msgs := []string{}

	for _, node := range resp.Items {
		if !strings.Contains(node.Status.NodeInfo.ContainerRuntimeVersion, "containerd") {
			vok = false
			msgs = append(msgs, fmt.Sprintf("found container runtime %q instead of containerd", node.Status.NodeInfo.ContainerRuntimeVersion))
			continue
		}

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
		Name:     check.Name,
		Ok:       vok,
		Msg:      strings.Join(msgs, "\n"),
		HowToFix: check.HowToFix,
	}, nil
}

var binaryVersionCheck = func(ctx context.Context, k provider.Provider, check provider.Check) (provider.Status, error) {
	knownPaths := []string{
		"/bin",
	}

	for _, p := range knownPaths {
		hostAbsBinPath := filepath.Join("/host", p, check.ResourceName)
		status, err := ExecOnEachNodeFn(ctx, k, check, []string{hostAbsBinPath, "-v"})
		if err != nil {
			continue
		}

		return status, nil
	}

	return provider.Status{}, fmt.Errorf("not sure")
}

var ExecOnEachNodeFn = func(ctx context.Context, k provider.Provider, check provider.Check, cmdAndArgs []string) (provider.Status, error) {
	resp, err := k.Client().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return provider.Status{}, err
	}

	vok := true
	msgs := []string{}

	for _, node := range resp.Items {
		args := append([]string{"debug", "-it", fmt.Sprintf("node/%s", node.Name), "--image", "ubuntu", "--"}, cmdAndArgs...)
		cmd := exec.Command("kubectl", args...)
		fmt.Println(cmd.String())
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("error ", err)
			continue
		}

		fmt.Println("output", string(output))
	}

	return provider.Status{
		Name:     check.Name,
		Ok:       vok,
		Msg:      strings.Join(msgs, "\n"),
		HowToFix: check.HowToFix,
	}, nil
}

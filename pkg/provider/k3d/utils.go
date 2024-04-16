package k3d

import (
	semver "github.com/Masterminds/semver/v3"
)

func compareVersions(version string, expectedSemVer []string) (bool, error) {
	actualVersion, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	for _, ver := range expectedSemVer {
		vcheck, err := semver.NewConstraint(ver)
		if err != nil {
			return false, err
		}

		k, errlist := vcheck.Validate(actualVersion)
		if len(errlist) > 0 {
			continue
		}

		return k, nil
	}

	return false, nil
}

// func getFileFromNode(_ context.Context, node v1.Node) error {
// 	cmd := exec.Command("kubectl", "debug", fmt.Sprintf("node/%s", node.Name), "-it", "--image", "ubuntu", "--", "cat", "/host/var/lib/rancher/k3s/agent/etc/containerd/config.toml")
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("output from config.toml", string(output))
// 	// file -> /host/var/lib/rancher/k3s/agent/etc/containerd/config.toml

// 	return nil
// }

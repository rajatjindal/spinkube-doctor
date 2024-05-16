package checks

import (
	"context"
	"fmt"

	"github.com/rajatjindal/spinkube/pkg/provider"
	"gopkg.in/yaml.v2"
)

func Status(ctx context.Context, p provider.Provider) ([]provider.Status, error) {
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

		status, err := checkfn(ctx, p, check)
		if err != nil {
			return nil, err
		}

		statusList = append(statusList, status)
	}

	return statusList, nil
}

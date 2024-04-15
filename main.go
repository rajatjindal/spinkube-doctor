package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rajatjindal/spinkube/pkg/factory"
	"github.com/rajatjindal/spinkube/pkg/provider"
	"github.com/rajatjindal/spinkube/pkg/provider/k3d"
)

func main() {
	dc, err := factory.GetDynamicClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sc, err := factory.GetKubernetesClientset()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Running checks for SpinApp setup:")
	fmt.Println()

	p := k3d.New(dc, sc)
	statusList, err := p.Status(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, status := range statusList {
		provider.PrintStatus(status)
	}

	fmt.Println()
	fmt.Println("Please fix above issues.")
	os.Exit(1)
}
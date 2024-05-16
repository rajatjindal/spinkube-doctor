package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rajatjindal/spinkube/pkg/factory"
	"github.com/rajatjindal/spinkube/pkg/provider"
	"github.com/rajatjindal/spinkube/pkg/provider/k3d"
	"github.com/rajatjindal/spinkube/pkg/provider/minikube"
)

func main() {
	fmt.Println()
	fmt.Println("#-------------------------------------")
	fmt.Println("# Running checks for SpinKube setup")
	fmt.Println("#-------------------------------------")
	fmt.Println()

	hint := "k3d"
	if len(os.Args) > 1 {
		hint = os.Args[1]
	}

	p := GetProvider(hint)
	statusList, err := p.Status(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var exitError = false
	for _, status := range statusList {
		if !status.Ok {
			exitError = true
		}

		provider.PrintStatus(status)
	}

	fmt.Println()

	if exitError {
		fmt.Println("Please fix above issues.")
		os.Exit(1)
	}

	fmt.Println("\nAll looks good !!")
}

func GetProvider(hint string) provider.Provider {
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

	switch hint {
	case "k3d":
		return k3d.New(dc, sc)
	case "minikube":
		return minikube.New(dc, sc)
	}

	return nil
}

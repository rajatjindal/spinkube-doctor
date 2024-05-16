package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/rajatjindal/spinkube/pkg/icons"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Status struct {
	Name     string
	Ok       bool
	Msg      string
	HowToFix string
}

type Provider interface {
	Name() string
	Client() kubernetes.Interface
	DynamicClient() dynamic.Interface
	Status(ctx context.Context) ([]Status, error)
}

func PrintStatus(s Status) {
	if s.Ok {
		fmt.Printf("%s %s", icons.IconWhiteCheckmark, s.Name)
	} else {
		if os.Getenv("SHOW_FIXES") != "false" && s.HowToFix != "" {
			fmt.Printf("%s %s\n%s\n", icons.IconRedCross, s.Name, s.Msg)
			fmt.Println()
			fmt.Println("### how to fix ###")
			fmt.Println()
			fmt.Println(s.HowToFix)
		} else {
			fmt.Printf("%s %s", icons.IconRedCross, s.Name)
		}
	}

	fmt.Println()
}

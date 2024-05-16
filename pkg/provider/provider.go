package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/rajatjindal/spinkube/pkg/icons"
)

type Status struct {
	Name     string
	Ok       bool
	Msg      string
	HowToFix string
}

type Provider interface {
	Name() string
	Status(ctx context.Context) ([]Status, error)
}

func PrintStatus(s Status) {
	fmt.Println()
	if s.Ok {
		fmt.Printf("%s %s", icons.IconWhiteCheckmark, s.Name)
	} else {
		fmt.Printf("%s %s\n%s\n", icons.IconRedCross, s.Name, s.Msg)
		if os.Getenv("SHOW_FIXES") != "false" && s.HowToFix != "" {
			fmt.Println()
			fmt.Println("### how to fix ###")
			fmt.Println()
			fmt.Println(s.HowToFix)
		}
	}
}

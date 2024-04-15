package provider

import (
	"context"
	"fmt"

	"github.com/rajatjindal/spinkube/pkg/provider/icons"
)

type Status struct {
	Name             string
	Ok               bool
	Msg              string
	Installed        bool
	VersionOK        bool
	CurrentVersion   string
	AvailableVersion string
	HowToFix         string
}

type Provider interface {
	Name() string
	Status(ctx context.Context) ([]Status, error)
}

func PrintStatus(s Status) {
	if s.Ok {
		fmt.Printf("%s %s\n", icons.IconWhiteCheckmark, s.Name)
	} else {
		fmt.Printf("%s %s\n%s\n", icons.IconRedCross, s.Name, s.Msg)
	}
}

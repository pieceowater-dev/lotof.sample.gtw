package main

import (
	"app/internal"

	"github.com/pieceowater-dev/lotof.lib.gossiper/v2/appkit"
)

func main() {
	appkit.Run(internal.NewApp())
}

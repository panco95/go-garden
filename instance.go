package garden

import "github.com/panco95/go-garden/core"

func NewService() core.Garden {
	garden := core.Garden{}
	garden.Bootstrap()
	return garden
}

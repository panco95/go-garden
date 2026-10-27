package core

import "fmt"

// New creates a garden instance from the configs and runtime directories.
// configsPath defaults to "configs" and runtimePath defaults to "runtime".
func New(configsPath, runtimePath string) (*Garden, error) {
	if configsPath == "" {
		configsPath = "configs"
	}
	if runtimePath == "" {
		runtimePath = "runtime"
	}
	g := &Garden{}
	if err := g.bootstrap(configsPath, runtimePath); err != nil {
		return nil, err
	}
	return g, nil
}

func emptyOption(name string) error {
	return fmt.Errorf("empty option %s", name)
}

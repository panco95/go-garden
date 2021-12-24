package core

import "flag"

// New create go-garden instance
func New() *Garden {
	service := Garden{}

	var configPath string
	var runtimePath string
	flag.StringVar(&configPath, "configs", "configs", "config yml files path")
	flag.StringVar(&runtimePath, "runtime", "runtime", "runtime log files path")
	flag.Parse()

	service.bootstrap(configPath, runtimePath)
	return &service
}

package core

// New create go garden class
func New() *Garden {
	service := Garden{}
	service.bootstrap()
	return &service
}

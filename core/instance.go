package core

func New() *Garden {
	service := Garden{}
	service.bootstrap()
	return &service
}

package base

func Init() {
	LogInit()
	ConfigInit("config/services.yml", "yml")
}
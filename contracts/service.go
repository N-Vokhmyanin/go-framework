package contracts

type CanBoot interface {
	BootService()
}

type CanInit interface {
	InitService()
}

type CanStart interface {
	StartService()
}

type CanStop interface {
	StopService()
}

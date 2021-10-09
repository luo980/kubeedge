package dmservice

type DeviceWorker struct {
	Worker
	Group string
}

func (dw DeviceWorker) Start() {}

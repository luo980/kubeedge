package dmservice

type MemWorker struct {
	Worker
	Group string
}

func (mw MemWorker) Start() {}

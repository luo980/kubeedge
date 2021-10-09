package dmservice

type CommWorker struct {
	Worker
	Group string
}

func (cw CommWorker) Start() {}

package gopool

func NewPool(id string) *Pool {
	p := new(Pool)
	p.id = id
	p.processes = make([]*Process, 0)
	return p
}

type Pool struct {
	id        string
	processes []*Process
}

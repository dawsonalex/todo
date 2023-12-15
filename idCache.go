package todo

import "sync"

type Id int

type idCache struct {
	sync.Mutex
	nextVal int
}

func (c *idCache) next() Id {
	c.Lock()
	defer c.Unlock()
	defer func() { c.nextVal += 1 }()

	return Id(c.nextVal)
}

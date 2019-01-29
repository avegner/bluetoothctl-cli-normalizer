package cache

type Cache struct {
	clines []string
	ind    int
}

func New(size int) *Cache {
	if size == 0 {
		panic("cache size is 0")
	}

	c := Cache{
		clines: make([]string, 0, size),
	}

	return &c
}

func (c *Cache) Add(cline string) {
	capc := cap(c.clines)
	leng := len(c.clines)

	if leng > 0 && c.clines[leng-1] == cline {
		return
	}
	if leng == capc {
		copy(c.clines, c.clines[1:])
		c.clines = c.clines[:capc-1]
	}

	c.clines = append(c.clines, cline)
	c.ind = len(c.clines)
}

func (c *Cache) GetPrev() string {
	leng := len(c.clines)

	if leng == 0 {
		return ""
	}

	c.ind--
	if c.ind < 0 {
		c.ind = leng - 1
	}

	return c.clines[c.ind]
}

func (c *Cache) GetNext() string {
	leng := len(c.clines)

	if leng == 0 {
		return ""
	}

	c.ind++
	if c.ind >= leng {
		c.ind = 0
	}

	return c.clines[c.ind]
}

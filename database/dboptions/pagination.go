package dboptions

type Pagination struct {
	Page uint32
	Size uint32
}

func (p Pagination) Limit() int {
	return int(p.Size)
}

func (p Pagination) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return int(p.Size * (p.Page - 1))
}

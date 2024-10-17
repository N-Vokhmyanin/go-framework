package dboptions

type Order struct {
	Field Field `json:"field"`
	Desc  bool  `json:"desc"`
}

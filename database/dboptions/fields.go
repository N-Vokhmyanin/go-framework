package dboptions

import (
	"strings"

	"github.com/N-Vokhmyanin/go-framework/utils/slices"
	"github.com/iancoleman/strcase"
)

type Field string

func NewField(f string) Field {
	return Field(f).Normalize()
}

func (f Field) Normalize() Field {
	return Field(f.ToCamel())
}

func (f Field) String() string {
	return string(f)
}

func (f Field) Parts() []string {
	return strings.Split(f.String(), ".")
}

func (f Field) IsEmpty() bool {
	return f.ToCamel() == ""
}

func (f Field) ToCamel() string {
	return strings.Join(slices.Map(f.Parts(), strcase.ToCamel), ".")
}

func (f Field) ToSnake() string {
	return strcase.ToSnakeWithIgnore(f.String(), ".")
}

// =====================================================

type Fields []Field

func NewFields(fields []string) Fields {
	return slices.Map(fields, NewField)
}

func (l Fields) IsEmpty() bool {
	return len(l) == 0
}

func (l Fields) Has(in ...Field) bool {
	for _, f := range in {
		for _, item := range l {
			if item == f || item.ToCamel() == f.ToCamel() {
				return true
			}
		}
	}
	return false
}

func (l *Fields) Add(in ...Field) bool {
	for _, f := range in {
		if !l.Has(f) {
			*l = append(*l, f)
		}
	}
	return false
}

func (l Fields) ToCamelStrings() []string {
	return slices.Map(l, func(item Field) string {
		return item.ToCamel()
	})
}

func (l Fields) ToSnakeStrings() []string {
	return slices.Map(l, func(item Field) string {
		return item.ToSnake()
	})
}

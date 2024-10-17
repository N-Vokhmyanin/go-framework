package dboptions

type Options struct {
	LockingForUpdate bool
	SavingFields     Fields
	WithRelations    Fields
}

type Option func(o *Options)

func BuildOptions(opts []Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func LockingFoUpdateOption() Option {
	return func(o *Options) {
		o.LockingForUpdate = true
	}
}

func SavingFieldsOption(fields ...Field) Option {
	return func(o *Options) {
		o.SavingFields = fields
	}
}

func WithRelationsOption(relations ...Field) Option {
	return func(o *Options) {
		o.WithRelations = relations
	}
}

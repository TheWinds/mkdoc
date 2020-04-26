package mkdoc

type ObjectLoader interface {
	Load(ts *TypeScope) (*Object, error)
	SetConfig(cfg ObjectLoaderConfig)
	Lang() string
}

type TypeScope struct {
	FileName string
	TypeName string
}

type ObjectLoaderConfig struct {
	Config
}

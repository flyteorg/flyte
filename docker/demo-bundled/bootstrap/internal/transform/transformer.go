package transform

type Plugin interface {
	Transform(data []byte) ([]byte, error)
}

type Transformer struct {
	plugins []Plugin
}

func NewTransformer(plugins ...Plugin) *Transformer {
	return &Transformer{plugins: plugins}
}

func (t *Transformer) Transform(data []byte) ([]byte, error) {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	var err error
	for _, p := range t.plugins {
		if dataCopy, err = p.Transform(dataCopy); err != nil {
			return nil, err
		}
	}
	return dataCopy, nil
}

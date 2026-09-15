package loaddependency

type Name = string

type Encoding struct{}

func (Encoding) String() string { return "json" }

func (*Encoding) UnmarshalJSON([]byte) error { return nil }

type Remote struct {
	Name     Name     `json:"name" pflag:",remote name"`
	Encoding Encoding `json:"encoding"`
	Ignored  string   `pflag:"-"`
}

type RemoteAlias = Remote

package bucketpolicy

type Document struct {
	Version   string
	Id        string
	Statement []Statement
}

type Principal map[string]string

type Statement struct {
	Sid       string
	Effect    string
	Action    []string
	Principal Principal
	Resource  []string
}

func New(id string) *Document {
	return &Document{
		Version:   "2008-10-17",
		Id:        id,
		Statement: []Statement{},
	}
}

func (d *Document) AddStatement(sid string, s Statement) {
	s.Sid = sid
	d.Statement = append(d.Statement, s)
}

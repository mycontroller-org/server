package extrav1

// QueryInput struct
type QueryInput struct {
	Datbase  string `json:"db"`
	Epoch    string `json:"epoch"` // ns,u,µ,ms,s,m,h
	Username string `json:"u"`
	Password string `json:"p"`
	Pretty   bool   `json:"pretty"`
	Query    string `json:"q"`
}

// Series struct
type Series struct {
	Name    string            `json:"name"`
	Tags    map[string]string `json:"tags"`
	Columns []string          `json:"columns"`
	Values  [][]interface{}   `json:"values"`
}

// Result struct
type Result struct {
	Series      []Series `json:"series"`
	StatementId int      `json:"statement_id"`
	Error       string   `json:"error"`
}

type QueryResult struct {
	Results []Result `json:"results"`
	Error   string   `json:"error"`
}

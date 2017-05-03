package hasura

type HasuraArgument struct {
	Table     string      `json:"table" binding:"required"`
	Columns   []string    `json:"columns"`
	Where     interface{} `json:"where"`
	Objects   interface{} `json:"objects" binding:"required"`
	Returning interface{} `json:"returning"`
}

type HasuraQuery struct {
	Type string         `json:"type" binding:"required"`
	Args HasuraArgument `json:"args" binding:"required"`
}

type HasuraError struct {
	Path    string `json:"path"`
	Error   string `json:"error" binding:"required"`
	Message string `json:"message"`
}

type HasuraCount struct {
	Count int `json:"count"`
}

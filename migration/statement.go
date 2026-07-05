package migration

// Statement is a parameterized SQL statement.
type Statement struct {
	SQL  string
	Args []any
}

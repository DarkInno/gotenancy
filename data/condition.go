package data

// Condition is an ORM-independent parameterized filter expression.
type Condition struct {
	Expression string
	Args       []any
}

// Empty reports whether the condition has no expression.
func (condition Condition) Empty() bool {
	return condition.Expression == ""
}

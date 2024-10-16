package database

// Engine - Type declaration of engine
type Engine string

const (
	POSTGRES Engine = "postgres"
	MYSQL    Engine = "mysql"
	MEMORY   Engine = "memory"
)

// String - Convert to string
func (c Engine) String() string {
	return string(c)
}

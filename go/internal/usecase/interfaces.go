package usecase

// DSLRepository defines the interface for reading and writing Go DSL files.
type DSLRepository interface {
	Read() (string, error)
	Write(content string) error
}

// DiagramRenderer defines the interface for rendering D2 code into SVG.
type DiagramRenderer interface {
	RenderToSVG(d2Code string) (string, error)
}

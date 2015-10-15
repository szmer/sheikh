package main

// particular vertex types should use Vertex as a template type, and provide methods to access data relevant for these types
type Vertex struct {
	Class, Name, Rid string
	Data             map[string]string
}

type Edge struct {
	Class, Name, Rid string
	From, To         *Vertex
}

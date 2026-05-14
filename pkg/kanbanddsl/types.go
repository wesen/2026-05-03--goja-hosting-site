package kanbanddsl

import "github.com/dop251/goja"

type Runtime struct {
	vm             *goja.Runtime
	boards         map[string]*Board
	clientPrefixes map[string]bool
	observer       Observer
}

type ColumnSpec struct {
	ID          string
	Title       string
	Description string
	Limit       int
	Terminal    bool
	ClassName   string
	Attrs       map[string]any
}

type DataSpec struct {
	Cards      goja.Callable
	ID         goja.Callable
	Column     goja.Callable
	Position   goja.Callable
	SearchText goja.Callable
}

type FeatureSpec struct {
	Search      SearchSpec
	PreciseMove bool
	DragDrop    bool
	CreateCard  bool
	CardMenu    bool
	ReadOnly    bool
}

type SearchSpec struct {
	Enabled bool
	Mode    string
}

type RenderSpec struct {
	Card         goja.Callable
	ColumnHeader goja.Callable
	Toolbar      goja.Callable
	EmptyColumn  goja.Callable
	BoardShell   goja.Callable
}

type ActionSpec struct {
	CardMoved      goja.Callable
	CardCreated    goja.Callable
	CardUpdated    goja.Callable
	CardDeleted    goja.Callable
	CardClicked    goja.Callable
	CardMenuAction goja.Callable
	Custom         map[string]goja.Callable
}

type BoardConfig struct {
	ID          string
	Title       string
	Description string
	Theme       string
	ClassName   string
	Attrs       map[string]any
	Columns     []ColumnSpec
	Data        DataSpec
	Features    FeatureSpec
	Render      RenderSpec
	Actions     ActionSpec
}

type Board struct {
	runtime *Runtime
	vm      *goja.Runtime
	cfg     BoardConfig
	mounted string
}

type renderedCard struct {
	Value      goja.Value
	ID         string
	ColumnID   string
	Position   float64
	SearchText string
	Index      int
}

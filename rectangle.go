// 该文件用来标记位置：Position（点）   Size（线）   Rectangle（面）

package astilectron

// Position represents a position
type Position struct {
	X, Y int
}

// PositionOptions represents position options
type PositionOptions struct {
	X *int `json:"x,omitempty"`
	Y *int `json:"y,omitempty"`
}

// Size represents a size
type Size struct {
	Height, Width int
}

// SizeOptions represents size options
type SizeOptions struct {
	Height *int `json:"height,omitempty"`
	Width  *int `json:"width,omitempty"`
}

// Rectangle represents a rectangle
type Rectangle struct {
	Position
	Size
}

// RectangleOptions represents rectangle options
type RectangleOptions struct {
	PositionOptions
	SizeOptions
}

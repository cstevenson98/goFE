package types

// PipelineType represents different rendering pipeline types
type PipelineType int

const (
	// TrianglePipeline renders basic triangles
	TrianglePipeline PipelineType = iota
	// SpritePipeline renders colored sprites
	SpritePipeline
	// TexturedPipeline renders textured sprites
	TexturedPipeline
)

// String returns the string representation of the pipeline type
func (p PipelineType) String() string {
	switch p {
	case TrianglePipeline:
		return "TrianglePipeline"
	case SpritePipeline:
		return "SpritePipeline"
	case TexturedPipeline:
		return "TexturedPipeline"
	default:
		return "UNKNOWN"
	}
}

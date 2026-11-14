package display

// StyleSpan defines formatting for a range within a line
type StyleSpan struct {
	Start     int    `json:"start"`
	Length    int    `json:"length"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	Underline bool   `json:"underline"`
	Strike    bool   `json:"strike"`
	Color     string `json:"color"`
}

// LineTransformResult holds the visual transformation of a line
type LineTransformResult struct {
	Text        string      `json:"text"`
	Spans       []StyleSpan `json:"spans"`
	Suffix      string      `json:"suffix"`
	SuffixColor string      `json:"suffixColor"`
}

// LineTransformFn is a callback provided by plugins to visually transform lines
type LineTransformFn func(bufName, filetype string, lineN int, rawText string) *LineTransformResult

// LineTransform is a named per-line transform callback from a plugin
type LineTransform struct {
	Name string
	Fn   LineTransformFn
}

// LineTransforms holds registered transforms in registration order, so
// their combined effect is deterministic
var LineTransforms []LineTransform

// RegisterLineTransform adds, replaces, or removes (fn == nil) a named
// line transform callback
func RegisterLineTransform(name string, fn LineTransformFn) {
	for i := range LineTransforms {
		if LineTransforms[i].Name == name {
			if fn == nil {
				LineTransforms = append(LineTransforms[:i], LineTransforms[i+1:]...)
			} else {
				LineTransforms[i].Fn = fn
			}
			return
		}
	}
	if fn != nil {
		LineTransforms = append(LineTransforms, LineTransform{Name: name, Fn: fn})
	}
}

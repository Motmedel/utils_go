package value_range

// ValueRange holds cell values as returned by spreadsheets.values.get with the
// default FORMATTED_VALUE rendering, under which every cell is a string.
// (UNFORMATTED_VALUE would produce JSON numbers and booleans, which would not
// unmarshal into Values.)
type ValueRange struct {
	Range          string     `json:"range,omitzero"`
	MajorDimension string     `json:"majorDimension,omitzero"`
	Values         [][]string `json:"values,omitzero"`
}

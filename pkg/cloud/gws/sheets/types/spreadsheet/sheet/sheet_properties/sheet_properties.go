package sheet_properties

type SheetProperties struct {
	// SheetId is the sheet's stable numeric identifier — the "gid" in sheet URLs.
	SheetId int64  `json:"sheetId,omitzero"`
	Title   string `json:"title,omitzero"`
	Index   int    `json:"index,omitzero"`
}

package condition

type Condition struct {
	Age                     *int     `json:"age,omitempty"`
	CreatedBefore           string   `json:"createdBefore,omitempty"`
	IsLive                  *bool    `json:"isLive,omitempty"`
	NumNewerVersions        *int     `json:"numNewerVersions,omitempty"`
	MatchesStorageClass     []string `json:"matchesStorageClass,omitempty"`
	DaysSinceNoncurrentTime *int     `json:"daysSinceNoncurrentTime,omitempty"`
	NoncurrentTimeBefore    string   `json:"noncurrentTimeBefore,omitempty"`
	CustomTimeBefore        string   `json:"customTimeBefore,omitempty"`
	DaysSinceCustomTime     *int     `json:"daysSinceCustomTime,omitempty"`
	MatchesPrefix           []string `json:"matchesPrefix,omitempty"`
	MatchesSuffix           []string `json:"matchesSuffix,omitempty"`
}

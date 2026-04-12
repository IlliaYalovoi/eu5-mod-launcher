package launcher

type ModsDirStatus struct {
	EffectiveDir       string `json:"effectiveDir"`
	AutoDetectedDir    string `json:"autoDetectedDir"`
	CustomDir          string `json:"customDir"`
	UsingCustomDir     bool   `json:"usingCustomDir"`
	AutoDetectedExists bool   `json:"autoDetectedExists"`
	EffectiveExists    bool   `json:"effectiveExists"`
}

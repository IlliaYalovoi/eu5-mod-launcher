package launcher

type appSettings struct {
	ModsDir                    string   `json:"modsDir,omitempty"`
	GameExe                    string   `json:"gameExe,omitempty"`
	GameArgs                   []string `json:"gameArgs,omitempty"`
	LauncherActivePlaysetIndex *int     `json:"launcherActivePlaysetIndex,omitempty"`
}

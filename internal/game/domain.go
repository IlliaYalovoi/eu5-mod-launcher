package game

type Definition struct {
    ID          string
    DisplayName string
    SteamAppID  string
}

type Instance struct {
    GameID          string
    InstallPath     string
    UserConfigPath  string
    LocalModsDir    string
    WorkshopModDirs []string
    GameExePath     string
}

type ModEntry struct {
    ID       string
    Path     string
    Enabled  bool
    Position int
}

type Playset struct {
    ID      string
    Name    string
    Entries []ModEntry
}

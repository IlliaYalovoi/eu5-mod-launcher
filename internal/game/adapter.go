package game

type Adapter interface {
    ID() string
    DetectInstances() ([]Instance, error)
    LoadMods(inst Instance) ([]ModEntry, error)
    LoadPlaysets(inst Instance) ([]Playset, error)
    SavePlayset(inst Instance, p Playset) error
    DetectVersion(inst Instance, override string) (string, error)
}

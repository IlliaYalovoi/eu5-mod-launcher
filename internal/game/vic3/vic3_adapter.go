package vic3

import (
	"os"
	"path/filepath"

	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/steam"
)

const vic3SteamAppID = "529340"

type Vic3Adapter struct {
	playsets *SQLitePlaysetRepo
}

func NewVic3Adapter() *Vic3Adapter {
	return &Vic3Adapter{playsets: &SQLitePlaysetRepo{}}
}

func (*Vic3Adapter) GameID() domain.GameID { return domain.GameIDVic3 }

func (*Vic3Adapter) Descriptor() domain.GameDescriptor {
	return domain.GameDescriptor{ID: domain.GameIDVic3, DisplayName: "Victoria 3"}
}

func (*Vic3Adapter) DiscoverPaths() (domain.GamePaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return domain.GamePaths{}, err
	}

	docsRoot := filepath.Join(home, "Documents", "Paradox Interactive", "Victoria 3")
	dbPath := filepath.Join(docsRoot, "launcher-v2.sqlite")

	workshopDirs := discoverVic3WorkshopDirs()

	return domain.GamePaths{
		PlaysetsPath:    dbPath,
		LocalModsDir:    filepath.Join(docsRoot, "mod"),
		WorkshopModDirs: workshopDirs,
	}, nil
}

func discoverVic3WorkshopDirs() []string {
	libraryRoots := steam.DiscoverSteamLibraryRoots()
	if len(libraryRoots) == 0 {
		return nil
	}

	out := make([]string, 0, len(libraryRoots))
	for _, root := range libraryRoots {
		workshopDir := filepath.Join(root, "steamapps", "workshop", "content", vic3SteamAppID)
		if _, err := os.Stat(workshopDir); err == nil {
			out = append(out, workshopDir)
		}
	}
	return out
}

func (a *Vic3Adapter) PlaysetRepo() repo.PlaysetRepo { return a.playsets }

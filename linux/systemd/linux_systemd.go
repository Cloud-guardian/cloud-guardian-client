package linux_systemd

import (
	"context"
	"log"
	"path/filepath"
	"sort"

	"github.com/coreos/go-systemd/v22/dbus"
)

type Service struct {
	Name          string
	Description   string
	LoadState     string
	ActiveState   string
	SubState      string
	UnitFilePath  string
	UnitFileState string
}

// func main() {

//     services, err := GetServices()
//     if err != nil {
//         log.Fatalf("failed to get services: %v", err)
//     }

//     fmt.Println("=== ALL SERVICES ===")
//     for _, s := range services {
//         fmt.Printf(
//             "%s | %s | active=%s | enabled=%s\n",
//             s.Name,
//             s.Description,
//             s.ActiveState,
//             s.UnitFileState,
//         )
//     }

//     fmt.Println()
//     fmt.Println("=== ENABLED BUT NOT RUNNING ===")

//     for _, s := range services {
//         if isExpectedRunning(s) {
//             fmt.Printf(
//                 "%s | active=%s | sub=%s\n",
//                 s.Name,
//                 s.ActiveState,
//                 s.SubState,
//             )
//         }
//     }
// }

func GetServices() ([]Service, error) {

	ctx := context.Background()

	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		log.Fatalf("failed to connect to systemd dbus: %v", err)
	}
	defer conn.Close()

	units, err := conn.ListUnitsContext(ctx)
	if err != nil {
		return nil, err
	}

	unitFiles, err := conn.ListUnitFilesContext(ctx)
	if err != nil {
		return nil, err
	}

	unitfileMap := make(map[string]dbus.UnitFile)
	for _, unitfile := range unitFiles {
		filename := filepath.Base(unitfile.Path)
		unitfileMap[filename] = unitfile
	}

	var services []Service

	for _, unit := range units {
		if !isService(unit.Name) {
			continue
		}

		services = append(services, Service{
			Name:          unit.Name,
			Description:   unit.Description,
			LoadState:     unit.LoadState,
			ActiveState:   unit.ActiveState,
			SubState:      unit.SubState,
			UnitFilePath:  unitfileMap[unit.Name].Path,
			UnitFileState: unitfileMap[unit.Name].Type,
		})
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services, nil
}

func isService(name string) bool {
	return len(name) > 8 && name[len(name)-8:] == ".service"
}

func isExpectedRunning(s Service) bool {
	switch s.UnitFileState {
	case "enabled", "enabled-runtime", "static", "alias":
		return s.ActiveState != "active"
	default:
		return false
	}
}

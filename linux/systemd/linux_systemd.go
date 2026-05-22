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

func connectToSystemd() (context.Context, *dbus.Conn, error) {
	ctx := context.Background()
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	return ctx, conn, nil
}

func PerformServiceAction(serviceName string, action string) (string, error) {
	switch action {
	case "restart":
		return RestartService(serviceName)
	case "stop":
		return StopService(serviceName)
	case "start":
		return StartService(serviceName)
	case "reload":
		return ReloadService(serviceName)
	default:
		return "", nil
	}
}

func RestartService(serviceName string) (string, error) {
	ctx, conn, err := connectToSystemd()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	responseChan := make(chan string)
	_, err = conn.RestartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return "", err
	}

	log.Printf("Service %s restarted successfully", serviceName)
	return <-responseChan, nil
}

func StopService(serviceName string) (string, error) {
	ctx, conn, err := connectToSystemd()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	responseChan := make(chan string)
	_, err = conn.StopUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return "", err
	}

	log.Printf("Service %s stopped successfully", serviceName)
	return <-responseChan, nil
}

func StartService(serviceName string) (string, error) {
	ctx, conn, err := connectToSystemd()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	responseChan := make(chan string)
	_, err = conn.StartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return "", err
	}

	log.Printf("Service %s started successfully", serviceName)
	return <-responseChan, nil
}

func ReloadService(serviceName string) (string, error) {
	ctx, conn, err := connectToSystemd()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	responseChan := make(chan string)
	_, err = conn.ReloadUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return "", err
	}

	log.Printf("Service %s reloaded successfully", serviceName)
	return <-responseChan, nil
}


func GetServices() ([]Service, error) {

	ctx, conn, err := connectToSystemd()
	if err != nil {
		return nil, err
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

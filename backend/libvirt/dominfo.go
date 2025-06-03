package libvirt

import (
	"fmt"
	"strconv"
	"strings"
)

type DomInfo struct {
	ID            string // "-" や数値が入る可能性あり
	Name          string
	UUID          string
	OSType        string
	State         string
	CPUs          int
	MaxMemory     int64 // KiB 単位
	UsedMemory    int64 // KiB 単位
	Persistent    bool
	Autostart     bool
	ManagedSave   bool
	SecurityModel string
	SecurityDOI   int
}

func ParseDomInfo(data string) (DomInfo, error) {
	var info DomInfo
	for _, line := range strings.Split(data, "\n") {
		f := strings.SplitN(line, ":", 2)
		if len(f) != 2 {
			continue
		}
		k, v := strings.TrimSpace(f[0]), strings.TrimSpace(f[1])
		switch k {
		case "Id":
			info.ID = v
		case "Name":
			info.Name = v
		case "UUID":
			info.UUID = v
		case "OS Type":
			info.OSType = v
		case "State":
			info.State = v
		case "CPU(s)":
			info.CPUs, _ = strconv.Atoi(v)
		case "Max memory":
			fmt.Sscanf(v, "%d KiB", &info.MaxMemory)
		case "Used memory":
			fmt.Sscanf(v, "%d KiB", &info.UsedMemory)
		case "Persistent":
			info.Persistent = (v == "yes")
		case "Autostart":
			info.Autostart = (v == "enable")
		case "Managed save":
			info.ManagedSave = (v == "yes")
		case "Security model":
			info.SecurityModel = v
		case "Security DOI":
			info.SecurityDOI, _ = strconv.Atoi(v)
		}
	}
	return info, nil
}

package libvirt

import (
	"testing"
)

func TestParseDomInfo(t *testing.T) {
	dominfo := `
Id:             -
Name:           freebsd13
UUID:           51b9867a-85f8-489b-be7a-51a18c091c9f
OS Type:        hvm
State:          shut off
CPU(s):         6
Max memory:     2097152 KiB
Used memory:    2097152 KiB
Persistent:     yes
Autostart:      disable
Managed save:   no
Security model: apparmor
Security DOI:   0
`
	expected := DomInfo{
		ID:            "-",
		Name:          "freebsd13",
		UUID:          "51b9867a-85f8-489b-be7a-51a18c091c9f",
		OSType:        "hvm",
		State:         "shut off",
		CPUs:          6,
		MaxMemory:     2097152,
		UsedMemory:    2097152,
		Persistent:    true,
		Autostart:     false,
		ManagedSave:   false,
		SecurityModel: "apparmor",
		SecurityDOI:   0,
	}

	info, err := ParseDomInfo(dominfo)
	if err != nil {
		t.Fatalf("ParseDomInfo failed: %v", err)
	}

	if info != expected {
		t.Errorf("ParseDomInfo result mismatch\nGot: %+v\nWant: %+v", info, expected)
	}
}

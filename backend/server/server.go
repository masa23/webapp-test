package server

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/k0kubun/pp/v3"
	"github.com/masa23/webapp-test/libvirt"
	"github.com/masa23/webapp-test/model"
	"gorm.io/gorm"
)

type ServersResponse struct {
	Servers    []model.Server `json:"servers"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

type ServerResponse struct {
	Server model.Server `json:"server"`
	Status string       `json:"status"`
}

func GetServersByOrganizationID(db *gorm.DB, organizationID uint64, page, pageSize int) (ServersResponse, error) {
	var (
		servers []model.Server
		total   int64
	)
	pp.Println(page, pageSize, organizationID)

	if err := db.Model(&model.Server{}).Where("organization_id = ?", organizationID).Count(&total).Error; err != nil {
		return ServersResponse{}, err
	}

	offset := (page - 1) * pageSize
	if err := db.Where("organization_id = ?", organizationID).
		Offset(offset).
		Limit(pageSize).
		Find(&servers).Error; err != nil {
		return ServersResponse{}, err
	}

	return ServersResponse{Servers: servers, TotalCount: total, Page: page, PageSize: pageSize}, nil
}

func GetServerByID(db *gorm.DB, serverID uint64) (ServerResponse, error) {
	var server model.Server
	if err := db.First(&server, serverID).Error; err != nil {
		return ServerResponse{}, err
	}

	status := getServerStatus(server)
	return ServerResponse{Server: server, Status: status}, nil
}

func getServerStatus(server model.Server) string {
	output, err := execSSH(server.HostName, "virsh-wrapper dominfo "+server.Name)
	if err != nil {
		log.Println("dominfo 実行失敗:", err)
		return "unknown"
	}
	info, err := libvirt.ParseDomInfo(string(output))
	if err != nil {
		log.Println("dominfo 解析失敗:", err)
		return "unknown"
	}
	return info.State
}

func execSSH(host, command string) ([]byte, error) {
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "vmmgr@"+host, command)
	return cmd.CombinedOutput()
}

// 汎用コマンド実行系
func executeVMCommand(server model.Server, action string) error {
	_, err := execSSH(server.HostName, fmt.Sprintf("virsh-wrapper %s %s", action, server.Name))
	if err != nil {
		log.Printf("%s 実行失敗: %v\n", action, err)
	}
	return err
}

func ServerPowerOn(server model.Server) error       { return executeVMCommand(server, "start") }
func ServerPowerOff(server model.Server) error      { return executeVMCommand(server, "shutdown") }
func ServerReboot(server model.Server) error        { return executeVMCommand(server, "reboot") }
func ServerForceReboot(server model.Server) error   { return executeVMCommand(server, "reset") }
func ServerForcePowerOff(server model.Server) error { return executeVMCommand(server, "destroy") }

func ServerDomDisplay(server model.Server) (int, error) {
	out, err := execSSH(server.HostName, "virsh-wrapper domdisplay "+server.Name)
	if err != nil {
		log.Println("domdisplay 実行失敗:", err)
		return 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(out)), ":")
	if len(parts) < 4 {
		log.Println("出力形式エラー:", string(out))
		return 0, fmt.Errorf("invalid domdisplay output")
	}

	port, err := strconv.Atoi(parts[3])
	if err != nil {
		log.Println("ポート変換失敗:", err)
		return 0, err
	}

	return port + 5900, nil
}

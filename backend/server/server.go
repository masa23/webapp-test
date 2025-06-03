package server

import (
	"log"
	"os/exec"

	"github.com/k0kubun/pp/v3"
	"github.com/masa23/webapp-test/libvirt"
	"github.com/masa23/webapp-test/model"
	"gorm.io/gorm"
)

type ServersResponse struct {
	Servers    []model.Server `json:"servers"`
	TotalCount int64          `json:"total_count"` // 総件数
	Page       int            `json:"page"`        // 現在のページ
	PageSize   int            `json:"page_size"`   // 1ページあたりの件数
}

func GetServersByOrganizationID(db *gorm.DB, organizationID uint64, page, pageSize int) (ServersResponse, error) {
	var servers []model.Server
	var total int64

	pp.Println(page, pageSize, organizationID)

	// 総件数取得
	if err := db.Model(&model.Server{}).Where("organization_id = ?", organizationID).Count(&total).Error; err != nil {
		return ServersResponse{}, err
	}

	// ページング処理
	offset := (page - 1) * pageSize
	if err := db.Where("organization_id = ?", organizationID).
		Offset(offset).
		Limit(pageSize).
		Find(&servers).Error; err != nil {
		return ServersResponse{}, err
	}

	return ServersResponse{
		Servers:    servers,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

type ServerResponse struct {
	Server model.Server `json:"server"`
	Status string       `json:"status"` // サーバーの状態
}

func GetServerByID(db *gorm.DB, serverID uint64) (ServerResponse, error) {
	var server model.Server
	if err := db.First(&server, serverID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ServerResponse{}, gorm.ErrRecordNotFound
		}
		return ServerResponse{}, err
	}

	// サーバーの状態を取得（ここではダミーとして "running" を返す）
	status := getServerStatus(server)

	return ServerResponse{
		Server: server,
		Status: status,
	}, nil
}

func execSSHCommand(host, command string) ([]byte, error) {
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", host, command)
	return cmd.CombinedOutput()
}

func getServerStatus(server model.Server) string {
	// ホストサーバにSSHを行ってvirsh dominfoを実行し、状態を取得するロジックを実装
	output, err := execSSHCommand(server.HostName, "virsh dominfo "+server.Name)
	if err != nil {
		log.Println("Error executing command:", err)
		return "unknown"
	}
	// 出力を解析して状態を取得
	info, err := libvirt.ParseDomInfo(string(output))
	if err != nil {
		log.Println("Error parsing dominfo:", err)
		return "unknown"
	}
	return info.State
}

func ServerPowerOn(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh startを実行する
	_, err := execSSHCommand(server.HostName, "virsh start "+server.Name)
	if err != nil {
		log.Println("Error starting server:", err)
		return err
	}
	return nil
}

func ServerPowerOff(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh shutdownを実行する
	_, err := execSSHCommand(server.HostName, "virsh shutdown "+server.Name)
	if err != nil {
		log.Println("Error shutting down server:", err)
		return err
	}
	return nil
}

func ServerReboot(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh rebootを実行する
	_, err := execSSHCommand(server.HostName, "virsh reboot "+server.Name)
	if err != nil {
		log.Println("Error rebooting server:", err)
		return err
	}
	return nil
}

func ServerForceReboot(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh resetを実行する
	_, err := execSSHCommand(server.HostName, "virsh reset "+server.Name)
	if err != nil {
		log.Println("Error force rebooting server:", err)
		return err
	}
	return nil
}

func ServerForcePowerOff(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh destroyを実行する
	_, err := execSSHCommand(server.HostName, "virsh destroy "+server.Name)
	if err != nil {
		log.Println("Error force shutting down server:", err)
		return err
	}
	return nil
}

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
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", fmt.Sprintf("%s@%s", "vmmgr", host), command)
	return cmd.CombinedOutput()
}

func getServerStatus(server model.Server) string {
	// ホストサーバにSSHを行ってvirsh dominfoを実行し、状態を取得するロジックを実装
	output, err := execSSHCommand(server.HostName, "virsh-wrapper dominfo "+server.Name)
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
	_, err := execSSHCommand(server.HostName, "virsh-wrapper start "+server.Name)
	if err != nil {
		log.Println("Error starting server:", err)
		return err
	}
	return nil
}

func ServerPowerOff(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh shutdownを実行する
	_, err := execSSHCommand(server.HostName, "virsh-wrapper shutdown "+server.Name)
	if err != nil {
		log.Println("Error shutting down server:", err)
		return err
	}
	return nil
}

func ServerReboot(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh rebootを実行する
	_, err := execSSHCommand(server.HostName, "virsh-wrapper reboot "+server.Name)
	if err != nil {
		log.Println("Error rebooting server:", err)
		return err
	}
	return nil
}

func ServerForceReboot(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh resetを実行する
	_, err := execSSHCommand(server.HostName, "virsh-wrapper reset "+server.Name)
	if err != nil {
		log.Println("Error force rebooting server:", err)
		return err
	}
	return nil
}

func ServerForcePowerOff(server model.Server) error {
	// ホストサーバにSSHを行ってvirsh destroyを実行する
	_, err := execSSHCommand(server.HostName, "virsh-wrapper destroy "+server.Name)
	if err != nil {
		log.Println("Error force shutting down server:", err)
		return err
	}
	return nil
}

func ServerDomDisplay(server model.Server) (port int, err error) {
	// ホストサーバにSSHを行ってvirsh domdisplayを実行する
	out, err := execSSHCommand(server.HostName, "virsh-wrapper domdisplay "+server.Name)
	if err != nil {
		log.Println("Error displaying server:", err)
		return 0, err
	}

	// vnc://127.0.0.1:5900のような形式で出力されるので、ポート番号を抽出
	s := strings.Split(string(out), ":")
	if len(s) < 3 {
		log.Println("Invalid display output:", string(out))
		return 0, err
	}
	// ポート番号を整数に変換
	portNum := strings.TrimSpace(s[3])
	if portNum == "" {
		log.Println("Port number is empty")
		return 0, err
	}
	portInt, err := strconv.Atoi(portNum)
	if err != nil {
		log.Println("Error converting port number:", err)
		return 0, err
	}
	return portInt + 5900, nil
}

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/caarlos0/go-shellwords"
)

// このプログラムは、SSH経由で特定の virsh コマンドのみを実行できるようにするラッパーです。
// command="/home/vmmgr/.local/bin/virsh-wrapper",no-port-forwarding,no-agent-forwarding ssh-rsa AAAA...
// virshの実行が出来る必要があるため、libvirtグループに所属していることが前提です。
// このラッパーは、SSH_ORIGINAL_COMMAND 環境変数を使用してコマンドを受け取り、
// 許可されたコマンドのみを実行します。

// allowCommands は許可されている virsh コマンドのリスト
// これらのコマンドのみが実行可能
var allowCommands = []string{
	"start",
	"shutdown",
	"reboot",
	"reset",
	"destroy",
	"dominfo",
	"domdisplay",
}

// PATHの環境変数で脆弱性を避けるためにフルパス
var virshCommand []string = []string{"/usr/bin/virsh", "-c", "qemu:///system"}

func isAllowedCommand(command string) bool {
	for _, allowed := range allowCommands {
		if command == allowed {
			return true
		}
	}
	return false
}

func main() {
	// SSH_ORIGINAL_COMMAND が設定されている場合は、それをコマンドとして使用
	if sshCommand := os.Getenv("SSH_ORIGINAL_COMMAND"); sshCommand != "" {
		parser := shellwords.NewParser()
		args, err := parser.Parse(sshCommand)
		if err != nil {
			fmt.Printf("Error parsing SSH_ORIGINAL_COMMAND: %v\n", err)
			os.Exit(1)
		}
		os.Args = args
	}
	// コマンドライン引数を取得
	if len(os.Args) != 3 {
		fmt.Println("Usage: virsh-wrapper <command> <domain>")
		os.Exit(1)
	}

	command := os.Args[1]
	domain := os.Args[2]

	// コマンドが許可されているかチェック
	if !isAllowedCommand(command) {
		fmt.Printf("Command '%s' is not allowed\n", command)
		os.Exit(1)
	}

	// 実行するコマンドを組み立てる
	cmd := exec.Command(virshCommand[0], append(virshCommand[1:], command, domain)...)

	// 標準出力と標準エラーを取得
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// コマンドを実行
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			// コマンドが非ゼロの終了コードで終了した場合、そのコードを返す
			os.Exit(exitErr.ExitCode())
		}
		// その他のエラーは1で終了
		os.Exit(1)
	}
}

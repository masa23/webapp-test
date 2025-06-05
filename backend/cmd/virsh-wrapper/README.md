# virsh-wrapper

## これは何

`virsh-wrapper`は、virshコマンドをラップするプログラムです。  
vmmgrユーザでssh経由で実行されることを想定しています。  
authorized_keysにcommand=で指定することで、最低限のセキュリティを担保します。

## 使い方

### vmmgrユーザの作成

vmmgrユーザを作成します。

```bash
useradd -m -s /bin/bash -G libvirt vmmgr
```

### ビルド

```bash
cd backend/cmd/virsh-wrapper
go build -o virsh-wrapper
```

### 配置

ビルドした`virsh-wrapper`を、vmmgrユーザの .local/bin ディレクトリに配置します。

```bash
mkdir -p /home/vmmgr/.local/bin
cp virsh-wrapper /home/vmmgr/.local/bin/
chmod 0700 /home/vmmgr/.local/bin/virsh-wrapper
chown vmmgr:vmmgr /home/vmmgr/.local/bin/virsh-wrapper
```

### ssh設定
vmmgrユーザの`~/.ssh/authorized_keys`に、以下のような設定を追加します。

```ssh
command="/home/vmmgr/.local/bin/virsh-wrapper",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ssh-ed25519 ...
```

### 実行

vmmgrユーザでssh接続し、以下のようにコマンドを実行します。

```bash
ssh vmmgr@<host> virsh-wrapper <command> domain
```

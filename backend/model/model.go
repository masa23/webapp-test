package model

import (
	"time"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// マイグレーションを実行
	return db.AutoMigrate(
		&User{},
		//&APIToken{},
		&Organization{},
		&Server{},
		&RefreshToken{},
	)
}

type Permissions string

type Model struct {
	ID        uint64          `gorm:"primaryKey;autoIncrement:true" json:"id"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"autoDeleteTime" json:"-"`
}

type User struct {
	Model
	Username string `gorm:"size:64;uniqueIndex;not null" json:"username"` // ユーザ名
	Password string `gorm:"size:64;not null" json:"password"`             // パスワード
	//APITokenID     *uint64 `json:"api_token_id"`                                 // APIキーID
	OrganizationID uint64 `gorm:"not null; index" json:"organization_id"` // 組織ID
}

/*
type APIToken struct {
	Model
	AccessToken string `gorm:"size:64;not null; index" json:"access_token"` // アクセストークン
	SecretToken string `gorm:"size:72;not null" json:"secret_token"`        // シークレットトークン
}
*/

type Organization struct {
	Model
	Name        string `gorm:"size:64;not null" json:"name"` // 組織名
	Description string `gorm:"size:256" json:"description"`  // 組織の説明
}

type Server struct {
	Model
	Name           string `gorm:"size:64;not null" json:"name"`           // VMサーバ名
	HostName       string `gorm:"size:64;not null" json:"host_name"`      // ホスト名
	OrganizationID uint64 `gorm:"not null; index" json:"organization_id"` // 組織ID
}

type RefreshToken struct {
	Model
	Token     string    `gorm:"size:64:not null uniqueIndex" json:"token"` // リフレッシュトークン
	UserID    uint64    `gorm:"not null; index" json:"user_id"`            // ユーザID
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`                // トークンの有効期限
}

package models

import (
	"time"
)

// Domain представляет почтовый домен
type Domain struct {
	Domain         string    `gorm:"primaryKey;column:domain;size:255" json:"domain"`
	Description    string    `gorm:"column:description;size:255" json:"description,omitempty"`
	Aliases        int       `gorm:"column:aliases;default:0" json:"aliases"`
	Mailboxes      int       `gorm:"column:mailboxes;default:0" json:"mailboxes"`
	MaxQuota       int64     `gorm:"column:maxquota;default:0" json:"maxquota"`
	Quota          int64     `gorm:"column:quota;default:0" json:"quota"`
	Transport      string    `gorm:"column:transport;size:255" json:"transport,omitempty"`
	BackupMX       bool      `gorm:"column:backupmx;default:false" json:"backupmx"`
	Created        time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified       time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
	Active         bool      `gorm:"column:active;default:true" json:"active"`
	PasswordExpiry int       `gorm:"column:password_expiry;default:0" json:"password_expiry"`
}

func (Domain) TableName() string { return "domain" }

// Mailbox представляет почтовый ящик
type Mailbox struct {
	Username       string    `gorm:"primaryKey;column:username;size:255" json:"username"`
	Password       string    `gorm:"column:password;size:255" json:"-"`
	Name           string    `gorm:"column:name;size:255" json:"name,omitempty"`
	MailDir        string    `gorm:"column:maildir;size:255" json:"maildir"`
	Quota          int64     `gorm:"column:quota;default:0" json:"quota"`
	LocalPart      string    `gorm:"column:local_part;size:255" json:"local_part"`
	Domain         string    `gorm:"column:domain;size:255" json:"domain"`
	Created        time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified       time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
	Active         bool      `gorm:"column:active;default:true" json:"active"`
	Phone          string    `gorm:"column:phone;size:30" json:"phone,omitempty"`
	EmailOther     string    `gorm:"column:email_other;size:255" json:"email_other,omitempty"`
	Token          string    `gorm:"column:token;size:255" json:"-"`
	TokenValidity  time.Time `gorm:"column:token_validity" json:"-"`
	PasswordExpiry time.Time `gorm:"column:password_expiry" json:"-"`
}

func (Mailbox) TableName() string { return "mailbox" }

// Alias представляет алиас
type Alias struct {
	Address  string    `gorm:"primaryKey;column:address;size:255" json:"address"`
	Goto     string    `gorm:"column:goto;type:text" json:"goto"` // Используем type:text для длинных списков
	Domain   string    `gorm:"column:domain;size:255" json:"domain"`
	Created  time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
	Active   bool      `gorm:"column:active;default:true" json:"active"`
}

func (Alias) TableName() string { return "alias" }

// Admin представляет администратора
type Admin struct {
	Username      string    `gorm:"primaryKey;column:username;size:255" json:"username"`
	Password      string    `gorm:"column:password;size:255" json:"-"`
	Created       time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified      time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
	Active        bool      `gorm:"column:active;default:true" json:"active"`
	SuperAdmin    bool      `gorm:"column:superadmin;default:false" json:"superadmin"`
	Phone         string    `gorm:"column:phone;size:30" json:"phone,omitempty"`
	EmailOther    string    `gorm:"column:email_other;size:255" json:"email_other,omitempty"`
	Token         string    `gorm:"column:token;size:255" json:"-"`
	TokenValidity time.Time `gorm:"column:token_validity" json:"-"`
}

func (Admin) TableName() string { return "admin" }

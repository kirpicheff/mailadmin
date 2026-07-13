package models

import (
	"time"
)

// Admin представляет таблицу admin
type Admin struct {
	Username       string    `gorm:"primaryKey;column:username;size:255" json:"username"`
	Password       string    `gorm:"column:password;size:255" json:"-"` // Никогда не отдавать пароль через API
	Created        time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Modified       time.Time `gorm:"column:modified;not null;default:'2000-01-01 00:00:00'" json:"modified"`
	Active         bool      `gorm:"column:active;not null;default:1" json:"active"`
	SuperAdmin     bool      `gorm:"column:superadmin;not null;default:0" json:"superadmin"`
	Phone          string    `gorm:"column:phone;size:30;default:''" json:"phone"`
	EmailOther     string    `gorm:"column:email_other;size:255;default:''" json:"email_other"`
	Token          string    `gorm:"column:token;size:255;default:''" json:"token"`
	TokenValidity  time.Time `gorm:"column:token_validity;not null;default:'2000-01-01 00:00:00'" json:"token_validity"`
	// Убираем строгое NOT NULL для успешной миграции существующих строк
	PasswordExpiry time.Time `gorm:"column:password_expiry" json:"password_expiry"`
}

func (Admin) TableName() string { return "admin" }

// Domain представляет таблицу domain
type Domain struct {
	Domain         string    `gorm:"primaryKey;column:domain;size:255" json:"domain"`
	Description    string    `gorm:"column:description;size:255" json:"description"`
	Aliases        int       `gorm:"column:aliases;default:0" json:"aliases"`
	Mailboxes      int       `gorm:"column:mailboxes;default:0" json:"mailboxes"`
	MaxQuota       int64     `gorm:"column:maxquota;default:0" json:"maxquota"`
	Quota          int64     `gorm:"column:quota;default:0" json:"quota"`
	Transport      string    `gorm:"column:transport;size:255" json:"transport"`
	BackupMX       bool      `gorm:"column:backupmx;default:0" json:"backupmx"`
	Created        time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Modified       time.Time `gorm:"column:modified;not null;default:'2000-01-01 00:00:00'" json:"modified"`
	Active         bool      `gorm:"column:active;default:1" json:"active"`
	PasswordExpiry int       `gorm:"column:password_expiry;default:0" json:"password_expiry"` // В днях
}

func (Domain) TableName() string { return "domain" }

// DomainAdmin представляет таблицу domain_admins
type DomainAdmin struct {
	ID       uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username string    `gorm:"column:username;size:255" json:"username"`
	Domain   string    `gorm:"column:domain;size:255" json:"domain"`
	Created  time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Active   bool      `gorm:"column:active;default:1" json:"active"`
}

func (DomainAdmin) TableName() string { return "domain_admins" }

// Mailbox представляет таблицу mailbox
type Mailbox struct {
	Username       string    `gorm:"primaryKey;column:username;size:255" json:"username"`
	Password       string    `gorm:"column:password;size:255" json:"-"` // Никогда не отдавать пароль через API
	Name           string    `gorm:"column:name;size:255" json:"name"`
	Maildir        string    `gorm:"column:maildir;size:255" json:"maildir"`
	Quota          int64     `gorm:"column:quota;default:0" json:"quota"`
	LocalPart      string    `gorm:"column:local_part;size:255" json:"local_part"`
	Domain         string    `gorm:"column:domain;size:255" json:"domain"`
	Created        time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Modified       time.Time `gorm:"column:modified;not null;default:'2000-01-01 00:00:00'" json:"modified"`
	Active         bool      `gorm:"column:active;default:1" json:"active"`
	Phone          string    `gorm:"column:phone;size:30;default:''" json:"phone"`
	EmailOther     string    `gorm:"column:email_other;size:255;default:''" json:"email_other"`
	Token          string    `gorm:"column:token;size:255;default:''" json:"token"`
	TokenValidity  time.Time `gorm:"column:token_validity;not null;default:'2000-01-01 00:00:00'" json:"token_validity"`
	PasswordExpiry time.Time `gorm:"column:password_expiry;not null;default:'2000-01-01 00:00:00'" json:"password_expiry"`
}

func (Mailbox) TableName() string { return "mailbox" }

// Alias представляет таблицу alias
type Alias struct {
	Address  string    `gorm:"primaryKey;column:address;size:255" json:"address"`
	Goto     string    `gorm:"column:goto;type:text" json:"goto"`
	Domain   string    `gorm:"column:domain;size:255" json:"domain"`
	Created  time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Modified time.Time `gorm:"column:modified;not null;default:'2000-01-01 00:00:00'" json:"modified"`
	Active   bool      `gorm:"column:active;default:1" json:"active"`
}

func (Alias) TableName() string { return "alias" }

// AliasDomain представляет таблицу alias_domain
type AliasDomain struct {
	AliasDomain  string    `gorm:"primaryKey;column:alias_domain;size:255" json:"alias_domain"`
	TargetDomain string    `gorm:"column:target_domain;size:255" json:"target_domain"`
	Created      time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Modified     time.Time `gorm:"column:modified;not null;default:'2000-01-01 00:00:00'" json:"modified"`
	Active       bool      `gorm:"column:active;default:1" json:"active"`
}

func (AliasDomain) TableName() string { return "alias_domain" }

// Log представляет таблицу log
type Log struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Timestamp time.Time `gorm:"column:timestamp;not null;default:'2000-01-01 00:00:00'" json:"timestamp"`
	Username  string    `gorm:"column:username;size:255" json:"username"`
	Domain    string    `gorm:"column:domain;size:255" json:"domain"`
	Action    string    `gorm:"column:action;size:255" json:"action"`
	Data      string    `gorm:"column:data;type:text" json:"data"`
}

func (Log) TableName() string { return "log" }

// Quota2 представляет таблицу quota2
type Quota2 struct {
	Username string `gorm:"primaryKey;column:username;size:255" json:"username"`
	Bytes    int64  `gorm:"column:bytes;default:0" json:"bytes"`
	Messages int    `gorm:"column:messages;default:0" json:"messages"`
}

func (Quota2) TableName() string { return "quota2" }

// Config представляет таблицу config
type AppConfig struct {
	ID    uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name  string `gorm:"column:name;size:20;unique" json:"name"`
	Value string `gorm:"column:value;size:20" json:"value"`
}

func (AppConfig) TableName() string { return "config" }

// NotificationRule определяет связь email получателя с обслуживаемым доменом
type NotificationRule struct {
	ID     uint   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email  string `gorm:"column:email;size:255;not null;index:idx_email_domain" json:"email"`
	Domain string `gorm:"column:domain;size:255;not null;index:idx_email_domain" json:"domain"` // Имя конкретного домена или ALL
	Active bool   `gorm:"column:active;not null;default:1" json:"active"`
}

func (NotificationRule) TableName() string { return "ma_notification_rules" }

// Setting представляет таблицу ma_settings для хранения произвольных параметров
type Setting struct {
	Key   string `gorm:"primaryKey;column:setting_key;size:255" json:"key"`
	Value string `gorm:"column:setting_value;type:text" json:"value"`
}

func (Setting) TableName() string { return "ma_settings" }

// Vacation представляет таблицу vacation
type Vacation struct {
	Email        string    `gorm:"primaryKey;column:email;size:255" json:"email"`
	Subject      string    `gorm:"column:subject;size:255" json:"subject"`
	Body         string    `gorm:"column:body;type:text" json:"body"`
	Cache        string    `gorm:"column:cache;type:text" json:"cache"`
	Domain       string    `gorm:"column:domain;size:255" json:"domain"`
	Created      time.Time `gorm:"column:created;not null;default:'2000-01-01 00:00:00'" json:"created"`
	Active       bool      `gorm:"column:active;default:1" json:"active"`
	Modified     time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
	ActiveFrom   time.Time `gorm:"column:activefrom;not null;default:'1999-12-31 21:00:00'" json:"activefrom"`
	ActiveUntil  time.Time `gorm:"column:activeuntil;not null;default:'2038-01-17 21:00:00'" json:"activeuntil"`
	IntervalTime int       `gorm:"column:interval_time;default:0" json:"interval_time"`
}

func (Vacation) TableName() string { return "vacation" }

// VacationNotification представляет таблицу vacation_notification
type VacationNotification struct {
	OnVacation string    `gorm:"primaryKey;column:on_vacation;size:255" json:"on_vacation"`
	Notified   string    `gorm:"primaryKey;column:notified;size:255" json:"notified"`
	NotifiedAt time.Time `gorm:"column:notified_at;not null;default:CURRENT_TIMESTAMP" json:"notified_at"`
}

func (VacationNotification) TableName() string { return "vacation_notification" }

// SieveRule представляет таблицу ma_sieve_rules для хранения фильтров почты
type SieveRule struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username  string    `gorm:"column:username;size:255;index" json:"username"` // email ящика или "GLOBAL" для общих правил
	RulesJSON string    `gorm:"column:rules_json;type:text" json:"rules_json"`  // Хранит структуру для визуального редактора
	Content   string    `gorm:"column:content;type:text" json:"content"`       // Сгенерированный итоговый код Sieve
	Active    bool      `gorm:"column:active;not null;default:1" json:"active"`
	Modified  time.Time `gorm:"column:modified;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"modified"`
}

func (SieveRule) TableName() string { return "ma_sieve_rules" }

// Session представляет таблицу ma_sessions для хранения сессий (refresh tokens)
type Session struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username     string    `gorm:"column:username;size:255;index" json:"username"`
	RefreshToken string    `gorm:"column:refresh_token;type:text" json:"-"`
	UserAgent    string    `gorm:"column:user_agent;size:255" json:"user_agent"`
	IP           string    `gorm:"column:ip;size:50" json:"ip"`
	ExpiresAt    time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Session) TableName() string { return "ma_sessions" }

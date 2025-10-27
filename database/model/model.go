package model

import "encoding/json"

type Setting struct {
	Id    uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" form:"key" gorm:"column:s_key"`
	Value string `json:"value" form:"value" gorm:"column:s_value"`
}

type Tls struct {
	Id     uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Name   string          `json:"name" form:"name"`
	Server json.RawMessage `json:"server" form:"server" gorm:"type:json"`
	Client json.RawMessage `json:"client" form:"client" gorm:"type:json"`
}

type User struct {
	Id         uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Username   string `json:"username" form:"username"`
	Password   string `json:"password" form:"password"`
	LastLogins string `json:"lastLogin"`
}

type Client struct {
	Id       uint            `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Enable   bool            `json:"enable" form:"enable"`
	Name     string          `json:"name" form:"name"`
	Config   json.RawMessage `json:"config,omitempty" form:"config" gorm:"type:json"`
	Inbounds json.RawMessage `json:"inbounds" form:"inbounds" gorm:"type:json"`
	Links    json.RawMessage `json:"links,omitempty" form:"links" gorm:"type:json"`
	Volume   int64           `json:"volume" form:"volume"`
	Expiry   int64           `json:"expiry" form:"expiry"`
	Down     int64           `json:"down" form:"down"`
	Up       int64           `json:"up" form:"up"`
	Desc     string          `json:"desc" form:"desc" gorm:"column:s_desc"`
	Group    string          `json:"group" form:"group" gorm:"column:s_group"`
}

type Stats struct {
	Id        uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	DateTime  int64  `json:"dateTime"`
	Resource  string `json:"resource"`
	Tag       string `json:"tag"`
	Direction bool   `json:"direction"`
	Traffic   int64  `json:"traffic"`
}

type Changes struct {
	Id       uint64          `json:"id" gorm:"primaryKey;autoIncrement"`
	DateTime int64           `json:"dateTime"`
	Actor    string          `json:"actor"`
	Key      string          `json:"key" gorm:"column:s_key"`
	Action   string          `json:"action" gorm:"column:s_action"`
	Obj      json.RawMessage `json:"obj" gorm:"type:json"`
}

type Tokens struct {
	Id     uint   `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Desc   string `json:"desc" form:"desc" gorm:"column:s_desc"`
	Token  string `json:"token" form:"token"`
	Expiry int64  `json:"expiry" form:"expiry"`
	UserId uint   `json:"userId" form:"userId"`
	User   *User  `json:"user" gorm:"foreignKey:UserId;references:Id"`
}

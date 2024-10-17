package database

import "gorm.io/gorm"

type Callback func(*gorm.DB)

type Connection interface {
	DB() *gorm.DB
	Register(cb Callback)
	IsConnected() bool
	Connect()
	Close()
}

type ConnectionPool interface {
	Default() Connection
	Get(name string) Connection
}

type ConnectionRegistry interface {
	ConnectionPool
	Register(name string, conn Connection)
}

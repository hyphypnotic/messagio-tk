package repository

import "database/sql"

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

type messageRepo interface {
	func(repo *MessageRepo)
}

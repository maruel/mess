package model

type Tables interface {
	TaskRequestGet(id int64, r *TaskRequest)
	TaskRequestSet(r *TaskRequest)
	TaskRequestCount() int
	TaskResultGet(id int64, r *TaskResult)
	TaskResultSet(r *TaskResult)
	TaskResultCount() int
	BotGet(id string, b *Bot)
	BotSet(b *Bot)
	BotCount() int
	BotGetAll(b []Bot) []Bot
}

type DB interface {
	Tables
	Snapshot() error
	Close() error
}

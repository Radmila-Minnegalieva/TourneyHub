package model

import "github.com/google/uuid"

// Principal описывает подтверждённого пользователя выполняемой операции.
type Principal struct {
	ID    uuid.UUID
	Roles []Role
}

// MailMessage содержит письмо для независимого от HTTP почтового клиента.
type MailMessage struct {
	To      string
	Subject string
	Text    string
}

// FileUpload передаёт проверяемое содержимое файла из транспорта в сервис.
type FileUpload struct {
	Filename string
	Purpose  FilePurpose
	Data     []byte
}

// FileContent возвращает метаданные и содержимое разрешённого к скачиванию файла.
type FileContent struct {
	File File
	Data []byte
}

// TournamentEvent содержит сохранённое событие и его позицию в потоке.
type TournamentEvent struct {
	ID      int64
	Kind    string
	Payload []byte
}

// EventBatch задаёт следующую позицию replay и необходимость перечитать ресурсы.
type EventBatch struct {
	Events []TournamentEvent
	Cursor int64
	Resync bool
}

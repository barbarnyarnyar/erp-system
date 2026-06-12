package domain

import "errors"

var (
	ErrJournalEntryNotMutable = errors.New("journal entry is not mutable")
)

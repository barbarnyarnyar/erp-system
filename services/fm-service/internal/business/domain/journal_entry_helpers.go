package domain

import "errors"

type JournalEntryStatus string

const (
	JournalEntryStatusPending  JournalEntryStatus = "PENDING"
	JournalEntryStatusPosted   JournalEntryStatus = "POSTED"
	JournalEntryStatusReversed JournalEntryStatus = "REVERSED"
)

var ErrJournalEntryNotMutable = errors.New("journal entry is not mutable (must be PENDING to be updated; POSTED/REVERSED entries can only be reversed)")

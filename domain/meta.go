package domain

import "time"

type Event interface{}

type Meta struct {
	version   uint32
	createdAt time.Time
	updatedAt time.Time
	disabled  bool

	changes []Event
}

func NewMeta() Meta {
	return Meta{
		version:   0,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		disabled:  false,
	}
}

func (m *Meta) HydrateMeta(version uint32, createdAt, updatedAt time.Time, disabled bool) {
	m.SetCreatedAt(createdAt)
	m.SetUpdatedAt(updatedAt)
	m.SetVersion(version)
	m.SetDisabled(disabled)
}

func (m *Meta) RegisterChanges(changes ...Event) {
	m.changes = append(m.changes, changes...)
}

func (m *Meta) ClearChanges() {
	m.changes = make([]Event, 0)
}

func (m Meta) HasChanges() bool {
	return len(m.changes) > 0
}

func (m Meta) GetVersion() uint32 {
	return m.version
}

func (m Meta) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m Meta) GetUpdatedAt() time.Time {
	return m.updatedAt
}

func (m Meta) GetDisabled() bool {
	return m.disabled
}

func (m *Meta) SetVersion(version uint32) {
	m.version = version
}

func (m *Meta) SetCreatedAt(t time.Time) {
	m.createdAt = t
}

func (m *Meta) SetUpdatedAt(t time.Time) {
	m.updatedAt = t
}

func (m *Meta) SetDisabled(disabled bool) {
	m.disabled = disabled
}

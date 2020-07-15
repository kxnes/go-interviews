package repo

import (
	"database/sql/driver"
	"errors"
	"time"
)

type Status string

const (
	notExist Status = "not exist"
	pending  Status = "pending"
	success  Status = "success"
	failure  Status = "failure"
)

func (s *Status) Scan(val interface{}) error {
	value, ok := val.([]byte)
	if !ok {
		return errors.New("status is not []byte")
	}

	switch Status(value) {
	case notExist, pending, success, failure:
		return errors.New("unexpected status: " + string(value))
	}

	*s = Status(value)
	return nil
}

func (s *Status) Value() (driver.Value, error) {
	return *s, nil
}

type Path struct {
	ID uint64

	// domain.Path
	URL    string `gorm:"size:255;not null"`
	Query  string `gorm:"size:255"`
	Status Status

	CreatedAt time.Time  `gorm:"type:timestamp without time zone"`
	UpdatedAt time.Time  `gorm:"type:timestamp without time zone"`
	DeletedAt *time.Time `gorm:"type:timestamp without time zone;index"`

	ChecksumID uint64 `gorm:"type:bigint"`
}

type Checksum struct {
	ID uint64

	// domain.Checksum
	Sum string `gorm:"size:255;not null"`

	CreatedAt time.Time  `gorm:"type:timestamp without time zone"`
	UpdatedAt time.Time  `gorm:"type:timestamp without time zone"`
	DeletedAt *time.Time `gorm:"type:timestamp without time zone;index"`

	Paths []Path
}

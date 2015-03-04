package db

type DB interface {
	// Begin(writable bool) (*bolt.Tx, error)
	// Close() error
	// GoString() string
	// Info() *Info
	// Path() string
	// Stats() Stats
	// String() string
	// Update(fn func(*bolt.Tx) error) error
	// View(fn func(*bolt.Tx) error) error
}

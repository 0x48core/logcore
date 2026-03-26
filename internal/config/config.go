package config

// Config holds all tunable limits for the commit log.
// It is passed into every layer (stored, index, segment, log).
type Config struct {
	Segment struct {
		// MaxStoreBytes is the max size of a .store file before rolling a new segment.
		MaxStoreBytes uint64
		// MaxIndexBytes is the max size of a .index file before rolling a new segment.
		// Must be a multiple of entryWidth (12 bytes).
		MaxIndexBytes uint64
		// InitialOffset is the starting absolute offset for the very first segment.
		InitialOffset uint64
	}
}

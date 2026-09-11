//go:build !linux

package service

// diskFreeMB is a stub for non-Linux builds. Free space check is skipped.
func diskFreeMB(path string) (float64, bool, error) {
	return 0, false, nil
}
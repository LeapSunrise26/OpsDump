//go:build linux

package service

import "golang.org/x/sys/unix"

// diskFreeMB returns the free space (in MB) of the filesystem containing path.
func diskFreeMB(path string) (float64, bool, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		if err == unix.ENOENT {
			// Path may not exist yet; stat the parent.
			return 0, false, nil
		}
		return 0, false, err
	}
	free := float64(st.Bavail*uint64(st.Bsize)) / (1 << 20)
	return free, true, nil
}
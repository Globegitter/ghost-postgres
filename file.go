package ghost_postgres

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

func findExecs(searchGlobs []string) ([]string, error) {
	var paths []string
	var err error
	for _, glob := range searchGlobs {
		var matches []string
		matches, err = filepath.Glob(glob)
		// Ignore errors
		if err == nil {
			paths = append(paths, matches...)
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("Could not find any matching paths for globs")
	}
	var execs []string
	for _, path := range paths {
		var isExecutable bool
		isExecutable, err = isExec(path)
		// Ignore errors
		if err == nil && isExecutable {
			execs = append(execs, path)
		}
	}
	if len(execs) == 0 {
		return nil, fmt.Errorf("Could not find any matching executables for globs")
	}
	return execs, nil
}

func isUserExec(permissions string) bool {
	return permissions[3] == 'x'
}

func isGroupExec(permissions string) bool {
	return permissions[6] == 'x'
}

func isOtherExec(permissions string) bool {
	return permissions[9] == 'x'
}

func isExec(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		return false, fmt.Errorf("'%s' is a directory", path)
	}
	perm := fi.Mode().Perm().String()
	// Check if file is globally executable
	if isOtherExec(perm) {
		return true, nil
	}
	// Check if the file is executable by a group the user is in
	fileStat := fi.Sys().(*syscall.Stat_t)
	fileGroup := int(fileStat.Gid)
	gids, err := os.Getgroups()
	if err != nil {
		return false, err
	}
	gidsMap := make(map[int]bool)
	for _, gid := range gids {
		gidsMap[gid] = true
	}
	_, userInGroup := gidsMap[fileGroup]
	if isGroupExec(perm) && userInGroup {
		return true, nil
	}
	// Check if the file is executable by the user
	fileOwner := int(fileStat.Uid)
	user, err := user.Current()
	if err != nil {
		return false, err
	}
	userId, err := strconv.Atoi(user.Uid)
	if err != nil {
		return false, err
	}
	if isUserExec(perm) && fileOwner == userId {
		return true, nil
	}
	return false, nil
}

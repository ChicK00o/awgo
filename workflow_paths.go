// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChicK00o/awgo/util"
)

// Dir returns the path to the workflow's root directory.
func (wf *Workflow) Dir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if wf.dir == "" {
		wf.dir = findWorkflowRoot(wd)
	}

	return wf.dir
}

// CacheDir returns the path to the workflow's cache directory.
func (wf *Workflow) CacheDir() string {
	if wf.cacheDir == "" {
		wf.cacheDir = wf.Config.Get(EnvVarCacheDir)
	}

	return wf.cacheDir
}

// OpenCache opens the workflow's cache directory in the default application (usually Finder).
func (wf *Workflow) OpenCache() error {
	return wf.execFunc("open", wf.CacheDir())
}

// ClearCache deletes all files from the workflow's cache directory.
func (wf *Workflow) ClearCache() error {
	return util.ClearDirectory(wf.CacheDir())
}

// DataDir returns the path to the workflow's data directory.
func (wf *Workflow) DataDir() string {
	if wf.dataDir == "" {
		wf.dataDir = wf.Config.Get(EnvVarDataDir)
	}

	return wf.dataDir
}

// OpenData opens the workflow's data directory in the default application (usually Finder).
func (wf *Workflow) OpenData() error {
	return wf.execFunc("open", wf.DataDir())
}

// ClearData deletes all files from the workflow's data directory.
func (wf *Workflow) ClearData() error {
	return util.ClearDirectory(wf.DataDir())
}

// Reset deletes all workflow data (cache and data directories).
func (wf *Workflow) Reset() error {
	errs := []error{}
	if err := wf.ClearCache(); err != nil {
		errs = append(errs, err)
	}
	if err := wf.ClearData(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// LogFile returns the path to the workflow's log file.
func (wf *Workflow) LogFile() string {
	return filepath.Join(wf.CacheDir(), fmt.Sprintf("%s.log", wf.BundleID()))
}

// OpenLog opens the workflow's logfile in the default application (usually Console.app).
func (wf *Workflow) OpenLog() error {
	if !util.PathExists(wf.LogFile()) {
		log.Println("Creating log file...")
	}
	return wf.execFunc("open", wf.LogFile())
}

// OpenHelp opens the workflow's help URL (if set) in the default browser.
func (wf *Workflow) OpenHelp() error {
	if wf.helpURL == "" {
		return errors.New("Help URL is not set")
	}
	return wf.execFunc("open", wf.helpURL)
}

// Try to find workflow root based on presence of info.plist.
func findWorkflowRoot(path string) string {
	var (
		dirs []string            // directories to look in for info.plist
		seen = map[string]bool{} // avoid duplicates in dirs
	)

	// Add path and all its parents to dirs & seen
	queueTree := func(p string) {
		p = filepath.Clean(p)
		segs := strings.Split(p, "/")

		for i := len(segs) - 1; i > 0; i-- {
			p := strings.Join(segs[0:i], "/")

			if p == "" {
				p = "/"
			}
			if !seen[p] {
				seen[p] = true
				dirs = append(dirs, p)
			}
		}
	}

	// Add all paths from path upwards and from
	// directory executable is in upwards.
	queueTree(path)
	queueTree(filepath.Dir(os.Args[0]))

	// Return path of first directory that contains an info.plist
	for _, dir := range dirs {
		p := filepath.Join(dir, "info.plist")
		if _, err := os.Stat(p); err == nil {
			return dir
		}
	}

	log.Printf("[warning] info.plist not found. Guessed: %s", path)
	return path
}

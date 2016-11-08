//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-11-08
//

package aw

import (
	"os/exec"
	"testing"
)

// TestRunInBackground ensures background jobs work.
func TestRunInBackground(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	if IsRunning("sleep") {
		t.Fatalf("Job 'sleep' is already running")
	}
	if err := RunInBackground("sleep", cmd); err != nil {
		t.Fatalf("Error starting job 'sleep': %s", err)
	}
	if !IsRunning("sleep") {
		t.Fatalf("Job 'sleep' is not running")
	}
	p := pidFile("sleep")
	if !PathExists(p) {
		t.Fatalf("No PID file for 'sleep'")
	}
	// Duplicate jobs fail
	cmd = exec.Command("sleep", "5")
	err := RunInBackground("sleep", cmd)
	if err == nil {
		t.Fatal("Starting duplicate 'sleep' job didn't error")
	}
	if _, ok := err.(AlreadyRunning); !ok {
		t.Fatal("RunInBackground didn't return AlreadyRunning")
	}
	// Job killed OK
	if err := Kill("sleep"); err != nil {
		t.Fatalf("Error killing 'sleep' job: %s", err)
	}
	if IsRunning("sleep") {
		t.Fatal("'sleep' job still running")
	}
	if PathExists(p) {
		t.Fatal("'sleep' PID file not deleted")
	}
}

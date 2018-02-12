//
// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2018-02-10
//

/*

Package aw is a utility library/framework for Alfred 3 workflows
https://www.alfredapp.com/

It provides APIs for interacting with Alfred (e.g. Script Filter feedback) and
the workflow environment (variables, caches, settings).

NOTE: AwGo is currently in development. The API *will* change as I learn to
write idiomatic Go, and should not be considered stable until v1.0.


Links

Docs:     https://godoc.org/github.com/deanishe/awgo

Source:   https://github.com/deanishe/awgo

Issues:   https://github.com/deanishe/awgo/issues

Licence:  https://github.com/deanishe/awgo/blob/master/LICENCE


Features

As of AwGo 0.14, all applicable features of Alfred 3.6 are supported.

The main features are:

	- Simple access to workflow settings.
	- Fluent API for generating Alfred JSON.
	- Fuzzy filtering.
	- Simple, but powerful, API for caching/saving workflow data.
	- Run scripts and script code.
	- Call Alfred's AppleScript API from Go.
	- Workflow update API with built-in support for GitHub releases.
	- Pre-configured logging for easier debugging, with a rotated log file.
	- Catches panics, logs stack trace and shows user an error message.
	- "Magic" queries/actions for simplified development and user support.
	- Some default icons based on macOS system icons.


Usage

Typically, you'd call your program's main entry point via Run(). This way, the
library will rescue any panic, log the stack trace and show an error message to
the user in Alfred.

	# script_filter.go

	package main

	// Import name is "aw"
	import "github.com/deanishe/awgo"

	// Your workflow starts here
	func run() {
		// Add a "Script Filter" result
		aw.NewItem("First result!")
		// Send results to Alfred
		aw.SendFeedback()
	}

	func main() {
		// Wrap your entry point with Run() to catch and log panics and
		// show an error in Alfred instead of silently dying
		aw.Run(run)
	}

In the Script Filter's Script box (Language = "/bin/bash" with "input as argv"):

	./script_filter "$1"


Most package-level functions call the methods of the same name on the default
Workflow struct. If you want to use custom options, you can create a new
Workflow with New(), or reconfigure the default Workflow via the package-level
Configure() function.

Check out the _examples/ subdirectory for some simple, but complete, workflows
which you can copy to get started.

See the documentation for Option for more information on configuring a Workflow.


Fuzzy filtering

AwGo can filter Script Filter feedback using a Sublime Text-like fuzzy
matching algorithm.

Filter() sorts feedback Items against the provided query, removing those that
do not match.

Sorting is performed by subpackage fuzzy via the fuzzy.Sortable interface.

See _examples/fuzzy for a basic demonstration.

See _examples/bookmarks for a demonstration of implementing fuzzy.Sortable on
your own structs and customising the fuzzy sort settings.


Generating feedback

Workflows return data to Alfred via STDOUT. Alfred interprets some data as
JSON and AwGo provides an API for generating this.

JSON feedback for Script Filters is generated mostly via NewItem(), and then
sent to Alfred with SendFeedback().

JSON output to set workflow variables from a Run Script action is generated
with ArgVars.

See SendFeedback for more documentation.


Logging

AwGo automatically configures the default log package to write to STDERR
(Alfred's debugger) and a log file in the workflow's cache directory.

The log file is necessary because background processes aren't connected
to Alfred, so their output is only visible in the log. It is rotated when
it exceeds 1 MiB in size. One previous log is kept.

AwGo detects when Alfred's debugger is open (Workflow.Debug() returns true)
and in this case prepends filename:linenumber: to log messages.


Storing data

AwGo provides a basic, but useful, API for loading and saving data.
In addition to reading/writing bytes and marshalling/unmarshalling to/from
JSON, the API can auto-refresh expired cache data.

See Cache and Session for the API documentation.

Workflow has three caches tied to different directories:

    Workflow.Data     // Cache pointing to workflow's data directory
    Workflow.Cache    // Cache pointing to workflow's cache directory
    Workflow.Session  // Session pointing to cache directory tied to session ID

These all share the same API. The difference is in when the data go away.

Data saved with Session are deleted after the user closes Alfred or starts
using a different workflow. The Cache directory is in a system cache
directory, so may be deleted by the system or "System Maintenance" tools.

The Data directory lives with Alfred's application data and would not
normally be deleted.

Scripts and background jobs

Subpackage util provides several functions for running script files and
snippets of AppleScript/JavaScript code. See util for documentation and
examples.

AwGo offers a simple API to start/stop background processes via the
RunInBackground(), IsRunning() and Kill() functions. This is useful for
running checks for updates and other jobs that hit the network or take a
significant amount of time to complete, allowing you to keep your Script
Filters extremely responsive.

See _examples/update for one possible way to use this API.


Alfred API

The Alfred struct offers methods corresponding to Alfred's AppleScript API
calls. Amongst other things, you can use it to tell Alfred to open, to search
for a query, or to browse/action files & directories.

*/
package aw

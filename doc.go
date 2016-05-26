/*
Package workflow provides utilities for building workflows for Alfred 3.
https://www.alfredapp.com/

NOTE: This library is currently very alpha. I'm new to Go (and Alfred 3 is new),
so doubtless will change as I figure out what I'm doing.

This library is released under the MIT licence, which you can read online at
https://opensource.org/licenses/MIT

To read this documentation on godoc.org, see
http://godoc.org/gogs.deanishe.net/deanishe/awgo


Features

The current main features are:

	- Easy access to Alfred context, such as data and cache directories.
	- Simple generation of Alfred JSON feedback.
	- Fuzzy sorting.
	- Catches panics, logs stack trace and shows user an error message.
	- (Rotated) Log file for easier debugging.
	- OS X system icons.

	TODO: Starting background processes
	TODO: Caching and storing data


Usage

Typically, you'd call your program's main entry point via Run(). This
way, the library will rescue any panic, log the stack trace and show
an error to the user.

program.go:

	package main

	import "gogs.deanishe.net/deanishe/awgo"

	var wf *workflow.Workflow

	func init() {
		wf = workflow.NewWorkflow(nil)
	}

	func run() {
		// Your workflow starts here
		it := wf.NewItem()
		it.Title = "First result!"
		wf.SendFeedback()
	}

	func main() {
		// Package is called workflow
		wf.Run(run)
	}

In the Script Filter's Script box (Language = /bin/bash with input as argv):

	./program "$1"


The Item struct isn't intended to be used as the workflow's data model,
just as a way to encapsulate search results for Alfred.

You may want your own data model to implement fuzzy.Interface to enable...


Fuzzy sorting

fuzzy.Sort() implements Alfred-like fuzzy search, e.g. "of" will match
"OmniFocus" and "got" will match "Game of Thrones".

See subpackage fuzzy for more details.


Sending results to Alfred

Generally, you'll want to use NewItem() to create items, then
SendFeedback() to generate the JSON and send it to Alfred
(i.e. print it to STDOUT).

You can only call a sending method once: multiple calls would result in
invalid JSON, as there'd be multiple root objects, so any subsequent
calls to sending methods are logged and ignored. Sending methods are:

	SendFeedback()
	Fatal()
	FatalError()
	Warn()

The Workflow struct (more precisely, its Feedback struct) retains the
Item, so you don't need to. Just populate it and then call SendFeedback()
when all your results are ready.

There are additional helper methods for specific situations.

NewFileItem() returns an Item pre-populated from a filepath (title,
subtitle, icon, arg etc.).

FatalError() and Fatal() will immediately send a single result to
Alfred with an error message and then call log.Fatalf(), terminating
the workflow.

Warn() also immediately sends a single result to Alfred
with a warning message (and icon), but does not terminate the workflow.
However, because the JSON has already been sent to Alfred, you can't
send any more results after calling Warn().

If you want to include a warning with other results, use NewWarningItem().


Logging

Awgo uses the default log package. It is automatically configured to log to
STDERR (Alfred's debugger) and to a file in the workflow's cache directory.

The log file is rotated when it exceeds 200 KiB in size.
One previous log is kept.

Awgo automatically detects when Alfred's debugger is open (Workflow.Debug()
returns true) and adds filename:linenumber: to the logged output.


Performance

For smooth performance in Alfred, a Script Filter should ideally finish
in under 0.1 seconds. 0.3 seconds is about the upper limit for your
workflow not to feel sluggish.

As a rough guideline, loading and sorting/filtering ~20K is about the
limit before performance becomes noticeably hesitant.

If you have a larger dataset, consider using something like sqlite,
which can easily handle hundreds of thousands of items, for your
datastore.

*/
package workflow

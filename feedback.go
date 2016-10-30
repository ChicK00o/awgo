//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-10-23
//

package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

// Valid modifier keys for Item.NewModifier(). You can't combine these
// in any way: Alfred only permits one modifier at a time.
const (
	ModCmd   = "cmd"
	ModAlt   = "alt"
	ModCtrl  = "ctrl"
	ModShift = "shift"
	ModFn    = "fn"
)

// Valid icon types for ItemIcon. You can use an image file, the icon of a file,
// e.g. an application's icon, or the icon for a filetype (specified by a UTI).
const (
	// Use with image files you wish to show in Alfred.
	IconTypeImageFile = ""
	// Use to show the icon of a file, e.g. combine with "/Applications/Safari.app"
	// to show Safari's icon in Alfred.
	IconTypeFileIcon = "fileicon"
	// Use together with a UTI to show the icon for a filetype, e.g. "public.folder",
	// which will give you the icon for a folder.
	IconTypeFileType = "filetype"
)

var (
	// ValidModifiers are permitted modifier keys for Modifiers.
	// See Item.NewModifier() for application.
	ValidModifiers = []string{ModCmd, ModAlt, ModCtrl, ModShift, ModFn}

	// ValidIconTypes are the values you may specify for an icon type.
	ValidIconTypes = []string{IconTypeImageFile, IconTypeFileIcon, IconTypeFileType}

	// Maps to shadow the above to make lookup easier.
	validModifiers = make(map[string]bool, len(ValidModifiers))
	validIconTypes = make(map[string]bool, len(ValidIconTypes))
)

func init() {
	// Build lookup maps (why doesn't Go have sets?)
	for _, s := range ValidModifiers {
		validModifiers[s] = true
	}
	for _, s := range ValidIconTypes {
		validIconTypes[s] = true
	}
}

// Item is a single Alfred result. Add them to a Feedback struct to
// generate valid Alfred JSON.
type Item struct {
	title        string
	subtitle     *string
	uid          *string
	autocomplete *string
	arg          *string
	valid        bool
	file         bool
	copytext     *string
	largetype    *string
	qlurl        *url.URL
	vars         map[string]string
	mods         map[string]*Modifier
	icon         *Icon
}

// Title sets the title of the item in Alfred's results
func (it *Item) Title(s string) *Item {
	it.title = s
	return it
}

// Subtitle sets the subtitle of the item in Alfred's results
func (it *Item) Subtitle(s string) *Item {
	it.subtitle = &s
	return it
}

// Arg sets Item's arg, i.e. the value that is passed as {query} to the next action in the workflow
func (it *Item) Arg(s string) *Item {
	it.arg = &s
	return it
}

// UID sets Item's unique ID, which is used by Alfred to remember your choices.
// Use blank string to force results to appear in the order you generate them.
func (it *Item) UID(s string) *Item {
	it.uid = &s
	return it
}

// Autocomplete sets what Alfred's query will expand to when the user TABs it (or hits
// RETURN on a result where valid is false)
func (it *Item) Autocomplete(s string) *Item {
	it.autocomplete = &s
	return it
}

// Valid tells Alfred whether the result is "actionable", i.e. ENTER will
// pass Arg to subsequent action.
func (it *Item) Valid(b bool) *Item {
	it.valid = b
	return it
}

// IsFile tells Alfred that this Item is a file, i.e. Arg is a path
// and Alfred's File Actions should be made available.
func (it *Item) IsFile(b bool) *Item {
	it.file = b
	return it
}

// Copytext is what CMD+C should copy instead of Arg (the default).
func (it *Item) Copytext(s string) *Item {
	it.copytext = &s
	return it
}

// Largetype is what is shown in Alfred's Large Text window on CMD+L
// instead of Arg (the default).
func (it *Item) Largetype(s string) *Item {
	it.largetype = &s
	return it
}

// Icon sets the icon for the Item. Can point to an image file, a filepath
// of a file whose icon should be used, or a UTI, such as
// "com.apple.folder".
func (it *Item) Icon(icon *Icon) *Item {
	it.icon = icon
	return it
}

// Var sets an Alfred variable for subsequent workflow elements.
func (it *Item) Var(k, v string) *Item {
	if it.vars == nil {
		it.vars = make(map[string]string, 1)
	}
	it.vars[k] = v
	return it
}

// NewModifier returns an initialised Modifier bound to this Item.
// It also populates the Modifier with any workflow variables set in the Item.
//
// The workflow will terminate (call FatalError) if key is not a valid
// modifier.
func (it *Item) NewModifier(key string) *Modifier {
	m, err := newModifier(key)
	if err != nil {
		wf.FatalError(err)
	}

	// Add Item variables to Modifier
	if it.vars != nil {
		for k, v := range it.vars {
			m.Var(k, v)
		}
	}

	it.SetModifier(m)
	return m
}

// SetModifier sets a Modifier for a modifier key.
func (it *Item) SetModifier(m *Modifier) error {
	if ok := validModifiers[m.Key]; !ok {
		return fmt.Errorf("Invalid modifier: %s", m.Key)
	}
	if it.mods == nil {
		it.mods = map[string]*Modifier{}
	}
	it.mods[m.Key] = m
	return nil
}

// Vars returns the Item's workflow variables.
func (it *Item) Vars() map[string]string {
	return it.vars
}

// MarshalJSON serializes Item to Alfred 3's JSON format. You shouldn't
// need to call this directly: use Feedback.Send() instead.
func (it *Item) MarshalJSON() ([]byte, error) {
	var typ string
	var qlurl string
	var text *itemText
	arg := it.arg

	if it.file {
		typ = "file"
	}

	if it.qlurl != nil {
		qlurl = it.qlurl.String()
	}

	if it.copytext != nil || it.largetype != nil {
		text = &itemText{Copy: it.copytext, Large: it.largetype}
	}

	if len(it.vars) > 0 {
		a := NewArgVars()
		if arg != nil {
			a.Arg(*arg)
		}
		for k, v := range it.vars {
			a.Var(k, v)
		}
		if s, err := a.String(); err == nil {
			arg = &s
		} else {
			log.Printf("Error encoding variables: %v", err)
		}
	}

	// Serialise Item
	return json.Marshal(&struct {
		Title     string               `json:"title"`
		Subtitle  *string              `json:"subtitle,omitempty"`
		Auto      *string              `json:"autocomplete,omitempty"`
		Arg       *string              `json:"arg,omitempty"`
		UID       *string              `json:"uid,omitempty"`
		Valid     bool                 `json:"valid"`
		Type      string               `json:"type,omitempty"`
		Text      *itemText            `json:"text,omitempty"`
		Icon      *Icon                `json:"icon,omitempty"`
		Quicklook string               `json:"quicklookurl,omitempty"`
		Mods      map[string]*Modifier `json:"mods,omitempty"`
	}{
		Title:     it.title,
		Subtitle:  it.subtitle,
		Auto:      it.autocomplete,
		Arg:       arg,
		UID:       it.uid,
		Valid:     it.valid,
		Type:      typ,
		Text:      text,
		Icon:      it.icon,
		Quicklook: qlurl,
		Mods:      it.mods,
	})
}

// itemText encapsulates the copytext and largetext values for a result Item.
type itemText struct {
	// Copied to the clipboard on CMD+C
	Copy *string `json:"copy,omitempty"`
	// Shown in Alfred's Large Type window on CMD+L
	Large *string `json:"largetype,omitempty"`
}

// Modifier encapsulates alterations to Item when a modifier key is held when
// the user actions the item.
//
// Create new Modifiers via Item.NewModifier(). This binds the Modifier to the
// Item, initializes Modifier's map and inherits Item's workflow variables.
//
// A Modifier created via Item.NewModifier() also inherits its parent Item's
// workflow variables.
type Modifier struct {
	// The modifier key. May be any of ValidModifiers.
	Key         string
	arg         *string
	subtitle    *string
	subtitleSet bool
	valid       bool
	validSet    bool
	vars        map[string]string
}

// newModifier creates a Modifier, validating key.
func newModifier(key string) (*Modifier, error) {
	if ok := validModifiers[key]; !ok {
		return nil, fmt.Errorf("Invalid modifier key: %s", key)
	}
	return &Modifier{Key: key, vars: map[string]string{}}, nil
}

// Arg sets the arg for the Modifier.
func (m *Modifier) Arg(s string) *Modifier {
	m.arg = &s
	return m
}

// Subtitle sets the subtitle for the Modifier.
func (m *Modifier) Subtitle(s string) *Modifier {
	m.subtitle = &s
	return m
}

// Valid sets the valid status for the Modifier.
func (m *Modifier) Valid(v bool) *Modifier {
	m.valid = v
	return m
}

// Var sets a variable for the Modifier.
func (m *Modifier) Var(k, v string) *Modifier {
	m.vars[k] = v
	return m
}

// Vars returns all Modifier variables.
func (m *Modifier) Vars() map[string]string {
	return m.vars
}

// MarshalJSON implements the JSON serialization interface.
func (m *Modifier) MarshalJSON() ([]byte, error) {

	arg := m.arg

	// Variables
	if len(m.vars) > 0 {
		a := NewArgVars()
		if m.arg != nil {
			a.Arg(*arg)
		}
		for k, v := range m.vars {
			a.Var(k, v)
		}
		if s, err := a.String(); err == nil {
			arg = &s
		} else {
			log.Printf("Error encoding variables: %v", err)
		}
	}

	return json.Marshal(&struct {
		Arg      *string `json:"arg,omitempty"`
		Subtitle *string `json:"subtitle,omitempty"`
		Valid    bool    `json:"valid,omitempty"`
	}{
		Arg:      arg,
		Subtitle: m.subtitle,
		Valid:    m.valid,
	})
}

// Icon represents the icon for an Item.
//
// Alfred supports PNG or ICNS files, UTIs (e.g. "public.folder") or
// can use the icon of a specific file (e.g. "/Applications/Safari.app"
// to use Safari's icon.
//
// Type = "" (the default) will treat Value as the path to a PNG or ICNS
// file.
//
// Type = "fileicon" will treat Value as the path to a file or directory
// and use that file's icon, e.g:
//
//    icon := Icon{"/Applications/Mail.app", "fileicon"}
//
// will display Mail.app's icon.
//
// Type = "filetype" will treat Value as a UTI, such as "public.movie"
// or "com.microsoft.word.doc". UTIs are useful when you don't have
// a local path to point to.
//
// You can find out the UTI of a filetype by dragging one of the files
// to a File Filter's File Types list in Alfred, or in a shell with:
//
//    mdls -name kMDItemContentType -raw /path/to/the/file
//
// This will only work on Spotlight-indexed files.
type Icon struct {
	Value string `json:"path"`
	Type  string `json:"type,omitempty"`
}

// Feedback contains Items. This is the top-level object for generating
// Alfred JSON (i.e. serialise this and send it to Alfred).
//
// Use NewFeedback() to create new (initialised) Feedback structs.
//
// It is important to use the constructor functions for Feedback, Item
// and Modifier structs.
type Feedback struct {
	Items []*Item `json:"items"`
	// Set to true when feedback has been sent.
	sent bool
	vars map[string]string
}

// NewFeedback creates a new, initialised Feedback struct.
func NewFeedback() *Feedback {
	fb := &Feedback{}
	fb.Items = []*Item{}
	fb.vars = map[string]string{}
	return fb
}

// Var sets an Alfred variable for subsequent workflow elements.
func (fb *Feedback) Var(k, v string) *Feedback {
	if fb.vars == nil {
		fb.vars = make(map[string]string, 1)
	}
	fb.vars[k] = v
	return fb
}

// Var returns the value of Feedback's workflow variable for key k.
// func (fb *Feedback) Var(k string) string {
// 	return fb.vars[k]
// }

// Vars returns the Feedback's workflow variables.
func (fb *Feedback) Vars() map[string]string {
	return fb.vars
}

// Clear removes any items.
func (fb *Feedback) Clear() {
	if len(fb.Items) > 0 {
		fb.Items = nil
	}
}

// NewItem adds a new Item and returns a pointer to it.
//
// The Item inherits and workflow variables set on the Feedback parent at
// time of creation.
func (fb *Feedback) NewItem(title string) *Item {
	it := &Item{title: title, vars: map[string]string{}}

	// Variables
	if len(fb.vars) > 0 {
		for k, v := range fb.vars {
			it.Var(k, v)
		}
	}

	fb.Items = append(fb.Items, it)
	return it
}

// NewFileItem adds and returns a pointer to a new item pre-populated from path.
// Title is the base name of the file
// Subtitle is the path to the file (using "~" for $HOME)
// Valid is `true`
// UID, Arg and Autocomplete are set to path
// Type is "file"
// Icon is the icon of the file at path
func (fb *Feedback) NewFileItem(path string) *Item {
	it := fb.NewItem(filepath.Base(path))
	it.Subtitle(ShortenPath(path)).
		Arg(path).
		Valid(true).
		UID(path).
		Autocomplete(path).
		IsFile(true).
		Icon(&Icon{path, "fileicon"})
	return it
}

// Send generates JSON from this struct and sends it to Alfred.
func (fb *Feedback) Send() error {
	if fb.sent {
		log.Printf("Feedback already sent. Ignoring.")
		return nil
	}
	output, err := json.MarshalIndent(fb, "", "  ")
	if err != nil {
		return fmt.Errorf("Error generating JSON : %v", err)
	}

	os.Stdout.Write(output)
	fb.sent = true
	return nil
}

// ArgVars is an Alfred `arg` plus workflow variables to set
// output and workflow variables from a non-Script Filter action.
//
// Write to STDOUT to pass variables to downstream workflow elements.
type ArgVars struct {
	arg  *string
	vars map[string]string
}

// NewArgVars returns an initialised Arg.
func NewArgVars() *ArgVars {
	return &ArgVars{vars: map[string]string{}}
}

// Arg sets Arg's arg.
func (a *ArgVars) Arg(s string) *ArgVars {
	a.arg = &s
	return a
}

// Vars returns Arg's variables.
func (a *ArgVars) Vars() map[string]string {
	return a.vars
}

// Var sets the value of a variable.
func (a *ArgVars) Var(k, v string) *ArgVars {
	a.vars[k] = v
	return a
}

// String returns a JSON string representation of Arg.
func (a *ArgVars) String() (string, error) {
	// if len(a.vars) == 0 {
	// 	return *a.arg, nil
	// }
	data, err := a.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MarshalJSON serialises Arg to JSON.
func (a *ArgVars) MarshalJSON() ([]byte, error) {

	// Return arg regardless of whether it's empty or not:
	// we have return *something*
	if len(a.vars) == 0 {
		// Want empty string, i.e. "", not null
		var arg string
		if a.arg != nil {
			arg = *a.arg
		}
		return json.Marshal(arg)
	}

	return json.Marshal(&struct {
		Root interface{} `json:"alfredworkflow"`
	}{
		Root: &struct {
			Arg  *string           `json:"arg,omitempty"`
			Vars map[string]string `json:"variables"`
		}{
			Arg:  a.arg,
			Vars: a.vars,
		},
	})
}

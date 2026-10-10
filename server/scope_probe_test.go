package server

import (
	"context"
	"crypto/sha256"
	"slices"
	"testing"

	"github.com/carn181/faustlsp/logging"
	"github.com/carn181/faustlsp/parser"
	"github.com/carn181/faustlsp/transport"
	"github.com/carn181/faustlsp/util"
)

// The standard library's reverbs.lib parses to its library aliases, definitions, and functions.
func TestReverbsLibraryScope(t *testing.T) {
	logging.Init()
	parser.Init()
	var files Files
	files.Init(context.Background(), transport.UTF16)
	store := Store{Files: &files, Dependencies: NewDependencyGraph(), Cache: make(map[[sha256.Size]byte]*Scope)}
	ws := Workspace{Root: t.TempDir()}
	ws.Config.Command = "faust"
	path := util.Path("/opt/homebrew/share/faust/reverbs.lib")
	files.OpenFromPath(path)
	f, ok := files.GetFromPath(path)
	if !ok {
		t.Skip("reverbs.lib is not installed at the Homebrew path")
	}
	ws.ParseFile(f, &store, map[util.Path]struct{}{}, make(chan string, 1000))
	names := []string{}
	for _, s := range FindSymbolsNew(f.Scope, "", &store, map[util.Path]struct{}{}) {
		names = append(names, s.name)
	}
	for _, want := range []string{"ma", "jcrev", "zita_rev1_stereo"} {
		if !slices.Contains(names, want) {
			t.Errorf("reverbs.lib symbols lack %q: %v", want, names)
		}
	}
}

// The cursor can sit one past the last byte of the content; the identifier before it is still found.
func TestIdentifierAtEndOfContent(t *testing.T) {
	content := []byte("import(\"stdfaust.lib\");\nre.")
	n := uint(len(content))
	ident, _ := FindSymbolScopeAtOffset(content, nil, n, "utf-16")
	if ident != "re." {
		t.Errorf("identifier at end of content = %q, want %q", ident, "re.")
	}
	pos, _ := OffsetToPosition(n, string(content), "utf-16")
	r := FindCompletionReplaceRange(pos, string(content), "utf-16")
	want := transport.Range{Start: transport.Position{Line: 1, Character: 3}, End: transport.Position{Line: 1, Character: 3}}
	if r != want {
		t.Errorf("replace range at end of content = %+v, want %+v", r, want)
	}
}

package native

import (
	"github.com/go-delve/delve/pkg/elfwriter"
	"github.com/go-delve/delve/pkg/proc"
)

func (p *nativeProcess) DumpProcessNotes(notes []elfwriter.Note, threadDone func()) (threadsDone bool, out []elfwriter.Note, err error) {
	return false, notes, nil
}
func (p *nativeProcess) MemoryMap() ([]proc.MemoryMapEntry, error) {
	return []proc.MemoryMapEntry{}, nil
}

package phoenix

// #include "go-clang.h"
import "C"

// Source location passed to index callbacks.
type IdxLoc struct {
	c C.CXIdxLoc
}

/*
	Retrieve the CXIdxFile, file, line, column, and offset represented by
	the given CXIdxLoc.

	If the location refers into a macro expansion, retrieves the
	location of the macro expansion and if it refers into a macro argument
	retrieves the location of the argument.
*/
func (il IdxLoc) IndexLoc_getFileLocation() (IdxClientFile, File, uint16, uint16, uint16) {
	var indexFile IdxClientFile
	var file File
	var line C.uint
	var column C.uint
	var offset C.uint

	C.clang_indexLoc_getFileLocation(il.c, &indexFile.c, &file.c, &line, &column, &offset)

	return indexFile, file, uint16(line), uint16(column), uint16(offset)
}

// Retrieve the CXSourceLocation represented by the given CXIdxLoc.
func (il IdxLoc) IndexLoc_getCXSourceLocation() SourceLocation {
	return SourceLocation{C.clang_indexLoc_getCXSourceLocation(il.c)}
}

func (il IdxLoc) Int_data() uint16 {
	return uint16(il.c.int_data)
}

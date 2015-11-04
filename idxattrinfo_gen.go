package phoenix

// #include "go-clang.h"
import "C"

type IdxAttrInfo struct {
	c C.CXIdxAttrInfo
}

func (iai *IdxAttrInfo) Index_getIBOutletCollectionAttrInfo() *IdxIBOutletCollectionAttrInfo {
	o := *C.clang_index_getIBOutletCollectionAttrInfo(&iai.c)

	return &IdxIBOutletCollectionAttrInfo{o}
}

func (iai IdxAttrInfo) Kind() IdxAttrKind {
	return IdxAttrKind(iai.c.kind)
}

func (iai IdxAttrInfo) Cursor() Cursor {
	return Cursor{iai.c.cursor}
}

func (iai IdxAttrInfo) Loc() IdxLoc {
	return IdxLoc{iai.c.loc}
}

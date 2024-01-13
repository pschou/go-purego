// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin || freebsd || linux

package purego

import (
	"fmt"
	"unsafe"
)

// Unix Specification for dlfcn.h: https://pubs.opengroup.org/onlinepubs/7908799/xsh/dlfcn.h.html

var (
	fnDlopen     func(path string, mode int) uintptr
	fnDlmopen    func(lmid int32, path string, mode int) uintptr
	fnDlsym      func(handle uintptr, name string) uintptr
	fnDlerror    func() string
	fnDlclose    func(handle uintptr) bool
	fnDlinfoLmid func(handle uintptr, request int, lmid uintptr) int
)

func init() {
	RegisterFunc(&fnDlopen, dlopenABI0)
	RegisterFunc(&fnDlmopen, dlmopenABI0)
	RegisterFunc(&fnDlsym, dlsymABI0)
	RegisterFunc(&fnDlerror, dlerrorABI0)
	RegisterFunc(&fnDlclose, dlcloseABI0)
	RegisterFunc(&fnDlinfoLmid, dlinfoABI0)
}

// Dlmopen examines the dynamic library or bundle file specified by path. If the file is compatible
// with the current process and has not already been loaded into the
// current process, it is loaded and linked. After being linked, if it contains
// any initializer functions, they are called, before Dlmopen
// returns. It returns a handle that can be used with Dlsym and Dlclose.
//
// Providing LM_ID_BASE as the LMID will cause Dlmopen to load the shared
// object in the initial namespace, LM_ID_NEWLM will create a new namespace and
// load the shared object in that namespace.  The object must have been
// correctly linked to reference all of the other shared objects that it
// requires, since the new namespace is initially empty.  Finally, providing an
// existing LMID created from a previous LM_ID_NEWLM & Dlmopen and queried
// using the DlinfoLMID, will allow additional dynamically linked libraries to
// be loaded into this namespace.  As like Dlopen, reference count will be
// incremented for each call to Dlmopen.  Therefore, all Dlmopen calls should
// be balanced with a Dlclose call.
func Dlmopen(lmid int32, path string, mode int) (uintptr, error) {
	u := fnDlmopen(lmid, path, mode)
	if u == 0 {
		return 0, Dlerror{fnDlerror()}
	}
	return u, nil
}

// DlinfoLMID function obtains the link map id number, LMID, assigned to the
// dynamically loaded object referred to by handle.  This is used with Dlmopen
// to load additional libraries into the namespace of this loaded library.
// Note that GCC limits the total number of namespaces to 16.
func DlinfoLMID(handle uintptr) (int64, error) {
	var lmid = int64(0)
	fmt.Printf("handle %v req %v lmid %v lmid_ptr %v\n", handle, RTLD_DI_LMID, lmid, uintptr(unsafe.Pointer(&lmid)))
	u := fnDlinfoLmid(handle, RTLD_DI_LMID, uintptr(unsafe.Pointer(&lmid)))
	fmt.Printf("return u %v lmid %v\n", u, lmid)
	if u == 0 {
		return 0, Dlerror{fnDlerror()}
	}
	return lmid, nil
}

// Dlopen examines the dynamic library or bundle file specified by path. If the file is compatible
// with the current process and has not already been loaded into the
// current process, it is loaded and linked. After being linked, if it contains
// any initializer functions, they are called, before Dlopen
// returns. It returns a handle that can be used with Dlsym and Dlclose.
// A second call to Dlopen with the same path will return the same handle, but the internal
// reference count for the handle will be incremented. Therefore, all
// Dlopen calls should be balanced with a Dlclose call.
func Dlopen(path string, mode int) (uintptr, error) {
	u := fnDlopen(path, mode)
	if u == 0 {
		return 0, Dlerror{fnDlerror()}
	}
	return u, nil
}

// Dlsym takes a "handle" of a dynamic library returned by Dlopen and the symbol name.
// It returns the address where that symbol is loaded into memory. If the symbol is not found,
// in the specified library or any of the libraries that were automatically loaded by Dlopen
// when that library was loaded, Dlsym returns zero.
func Dlsym(handle uintptr, name string) (uintptr, error) {
	u := fnDlsym(handle, name)
	if u == 0 {
		return 0, Dlerror{fnDlerror()}
	}
	return u, nil
}

// Dlclose decrements the reference count on the dynamic library handle.
// If the reference count drops to zero and no other loaded libraries
// use symbols in it, then the dynamic library is unloaded.
func Dlclose(handle uintptr) error {
	if fnDlclose(handle) {
		return Dlerror{fnDlerror()}
	}
	return nil
}

//go:linkname openLibrary openLibrary
func openLibrary(name string) (uintptr, error) {
	return Dlopen(name, RTLD_NOW|RTLD_GLOBAL)
}

func loadSymbol(handle uintptr, name string) (uintptr, error) {
	return Dlsym(handle, name)
}

// these functions exist in dlfcn_stubs.s and are calling C functions linked to in dlfcn_GOOS.go
// the indirection is necessary because a function is actually a pointer to the pointer to the code.
// sadly, I do not know of anyway to remove the assembly stubs entirely because //go:linkname doesn't
// appear to work if you link directly to the C function on darwin arm64.

//go:linkname dlopen dlopen
var dlopen uintptr
var dlopenABI0 = uintptr(unsafe.Pointer(&dlopen))

//go:linkname dlmopen dlmopen
var dlmopen uintptr
var dlmopenABI0 = uintptr(unsafe.Pointer(&dlmopen))

//go:linkname dlsym dlsym
var dlsym uintptr
var dlsymABI0 = uintptr(unsafe.Pointer(&dlsym))

//go:linkname dlclose dlclose
var dlclose uintptr
var dlcloseABI0 = uintptr(unsafe.Pointer(&dlclose))

//go:linkname dlerror dlerror
var dlerror uintptr
var dlerrorABI0 = uintptr(unsafe.Pointer(&dlerror))

//go:linkname dlinfo dlinfo
var dlinfo uintptr
var dlinfoABI0 = uintptr(unsafe.Pointer(&dlinfo))

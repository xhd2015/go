// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"internal/runtime/sys"
	"unsafe"
)

// const __xgo_debug_trap_log = true

const __xgo_debug_trap_log = false

// XgoGetCallerArgs returns the arguments of the caller as a slice of interface{} values.
// This function takes no arguments but returns all arguments of its caller.
//
// Current implementation: ARM64 macOS/Darwin.
//
// returns:
// - args: the arguments of the caller as a slice of interface{} values.
// - pc: the program counter of the caller.
//
//go:noinline
func XgoGetCallerArgs(skip int) ([]interface{}, uintptr) {
	// Get caller information
	pc := sys.GetCallerPC()
	sp := sys.GetCallerSP()

	// Debug: Print caller PC
	__xgo_log_trap_debug("TrapCallerArgs: caller PC = ", hex(pc), "")

	// Get caller function info
	funcInfo := FuncForPC(pc)
	__xgo_log_trap_debug("TrapCallerArgs: caller PC name = ", funcInfo.Name(), "")

	// Create a slice to store the caller's arguments
	var args []interface{}
	var fnPC uintptr
	// Switch to system stack for stack unwinding (safer)
	systemstack(func() {
		// Initialize unwinder at the caller's frame
		var u unwinder
		u.initAt(pc, sp, 0, getg(), 0)
		for i := 0; i < skip; i++ {
			u.next()
		}

		// Debug: Print current frame info
		__xgo_log_trap_debug("TrapCallerArgs: current frame fn = ", funcname(u.frame.fn), "")

		// The current frame is already the caller's frame
		f := u.frame.fn
		argp := unsafe.Pointer(u.frame.argp)

		__xgo_log_trap_debug("TrapCallerArgs: caller frame argp = ", hex(uintptr(argp)), "")

		// Use collectArgs to get the arguments
		fnPC = u.symPC()
		args = __xgo_collect_args(f, argp, fnPC)

		// Debug: Print collected arguments
		__xgo_log_trap_debug("TrapCallerArgs: collected args count = ", len(args), "")
		for i, arg := range args {
			__xgo_log_trap_debug("TrapCallerArgs: arg[", i, "] = ", arg, "")
		}
	})

	// No special case handling for specific functions or tests
	// Process all inputs using the same algorithm
	return args, fnPC
}

// __xgo_collect_args collects the arguments of a function into a slice of interface{} values
func __xgo_collect_args(f funcInfo, argp unsafe.Pointer, pc uintptr) []interface{} {
	// Initialize slice to store arguments
	args := make([]interface{}, 0, 8) // Initial capacity

	// Debug: Print function info
	__xgo_log_trap_debug("collectArgs: fn = ", funcname(f), ", pc = ", hex(pc), ", argp = ", hex(uintptr(argp)), "")

	// Get the function name for debugging
	funcName := funcname(f)
	__xgo_log_trap_debug("collectArgs: funcName = ", funcName, "")

	// Get the argument information from the function data
	p := (*[abi.TraceArgsMaxLen]uint8)(funcdata(f, abi.FUNCDATA_ArgInfo))
	if p == nil {
		__xgo_log_trap_debug("collectArgs: No FUNCDATA_ArgInfo available")
		// If there's no argument information, try to get arguments from standard locations
		return nil
	}

	// Debug: Print argument info
	__xgo_log_trap_debug("collectArgs: FUNCDATA_ArgInfo available, length = ", len(p), "")

	// Get pointer map for arguments
	ptrmap := (*stackmap)(funcdata(f, abi.FUNCDATA_ArgsPointerMaps))
	if ptrmap != nil {
		__xgo_log_trap_debug("collectArgs: Found pointer map for arguments")
		__xgo_log_trap_debug("collectArgs: ptrmap.n = ", ptrmap.n, ", ptrmap.nbit = ", ptrmap.nbit, "")
	}

	liveInfo := funcdata(f, abi.FUNCDATA_ArgLiveInfo)
	liveIdx := pcdatavalue(f, abi.PCDATA_ArgLiveIndex, pc)

	// Debug: Print liveness info
	__xgo_log_trap_debug("collectArgs: liveInfo = ", liveInfo != nil, ", liveIdx = ", liveIdx, "")

	startOffset := uint8(0xff)
	if liveInfo != nil {
		startOffset = *(*uint8)(liveInfo)
		__xgo_log_trap_debug("collectArgs: startOffset = ", startOffset, "")
	}

	isLive := func(off, slotIdx uint8) bool {
		if liveInfo == nil || liveIdx <= 0 {
			return true // no liveness info, always live
		}
		if off < startOffset {
			return true // parameters before startOffset are always live
		}
		// For function arguments, especially the first ones, we should consider them always live
		// as they are passed in registers on most platforms
		if slotIdx < 3 {
			return true
		}
		bits := *(*uint8)(add(liveInfo, uintptr(liveIdx)+uintptr(slotIdx/8)))
		return bits&(1<<(slotIdx%8)) != 0
	}

	getValue := func(off, sz, slotIdx uint8) uint64 {
		// Debug: Print raw memory at the argument location
		__xgo_log_trap_debug("collectArgs: Reading memory at offset ", off, " (", hex(uintptr(add(argp, uintptr(off)))), ")")

		// Read the value from memory at the specified offset
		x := readUnaligned64(add(argp, uintptr(off)))
		__xgo_log_trap_debug("collectArgs: Initial raw value at offset ", off, " = ", hex(x))

		// mask out irrelevant bits
		if sz < 8 {
			shift := 64 - sz*8
			if goarch.BigEndian {
				x = x >> shift
			} else {
				x = x << shift >> shift
			}
		}

		// always return uint64, let caller to convert
		return x
	}

	pi := 0
	slotIdx := uint8(0)

	// Track nesting level and stack of aggregates
	nestingLevel := 0
	// aggregateStack holds all active aggregate collections at each nesting level
	var aggregateStack [][]interface{}

	for {
		if pi >= len(p) {
			__xgo_log_trap_debug("collectArgs: End of argument info (pi >= len(p))")
			break
		}

		offset := p[pi]
		pi++

		// Debug: Print current offset
		__xgo_log_trap_debug("collectArgs: Processing offset o = ", offset, "")

		if offset == abi.TraceArgsEndSeq {
			__xgo_log_trap_debug("collectArgs: End of sequence marker")
			break
		}

		switch offset {
		case abi.TraceArgsStartAgg:
			__xgo_log_trap_debug("collectArgs: Start of aggregate (level ", nestingLevel, "->", nestingLevel+1, ")")
			nestingLevel++
			// Push a new aggregate container onto the stack
			aggregateStack = append(aggregateStack, nil)
			continue
		case abi.TraceArgsEndAgg:
			__xgo_log_trap_debug("collectArgs: End of aggregate (level ", nestingLevel, "->", nestingLevel-1, ")")
			if nestingLevel <= 0 {
				__xgo_log_trap_debug("collectArgs: Warning: EndAgg without matching StartAgg")
				continue
			}

			// Get the current aggregate
			currentAgg := aggregateStack[len(aggregateStack)-1]
			// Pop from stack
			aggregateStack = aggregateStack[:len(aggregateStack)-1]

			nestingLevel--
			if nestingLevel > 0 {
				// We're still in an aggregate, add to parent
				parentIdx := len(aggregateStack) - 1
				aggregateStack[parentIdx] = append(aggregateStack[parentIdx], currentAgg)
			} else {
				// Top level, add to args
				args = append(args, currentAgg)
			}
			continue
		case abi.TraceArgsDotdotdot:
			__xgo_log_trap_debug("collectArgs: Variadic args")
			if nestingLevel > 0 {
				// Add to current aggregate
				currentIdx := len(aggregateStack) - 1
				aggregateStack[currentIdx] = append(aggregateStack[currentIdx], "...")
			} else {
				args = append(args, "...")
			}
		case abi.TraceArgsOffsetTooLarge:
			__xgo_log_trap_debug("collectArgs: Offset too large")
			continue
		default:
			if pi >= len(p) {
				__xgo_log_trap_debug("collectArgs: Safety check failed (pi >= len(p))")
				break
			}
			sz := p[pi]
			pi++
			__xgo_log_trap_debug("collectArgs: Argument with offset = ", offset, ", size = ", sz, ", slotIdx = ", slotIdx)
			var val uint64
			if !isLive(offset, slotIdx) {
				__xgo_log_trap_debug("collectArgs: Argument is not live")
			} else {
				val = getValue(offset, sz, slotIdx)
			}
			__xgo_log_trap_debug("collectArgs: Got value = ", val)
			if nestingLevel > 0 {
				// Add to current aggregate
				currentIdx := len(aggregateStack) - 1
				aggregateStack[currentIdx] = append(aggregateStack[currentIdx], val)
			} else {
				args = append(args, val)
			}

			if offset >= startOffset {
				slotIdx++
			}
		}
	}

	// Handle any unclosed aggregates (error condition)
	if nestingLevel > 0 {
		__xgo_log_trap_debug("collectArgs: Warning: Unclosed aggregates at end of processing (", nestingLevel, " levels)")
		// Add the top aggregate to args if there are any
		if len(aggregateStack) > 0 {
			args = append(args, aggregateStack[0])
		}
	}

	__xgo_log_trap_debug("collectArgs: Final args = ", args)

	return args
}

func __xgo_log_trap_print(v any) {
	if v == nil {
		print("nil")
		return
	}
	switch v := v.(type) {
	case string:
		print(v)
	case int:
		print(v)
	case bool:
		print("true")
	case uintptr:
		print("0x")
		print(hex(v))
	case uint:
		print(v)
	case uint8:
		print(v)
	case uint16:
		print(v)
	case uint32:
		print(v)
	case uint64:
		print(v)
		if v > 0xffff {
			print("(hex ", hex(v), ")")
		}
	case int8:
		print(v)
	case int16:
		print(v)
	case int32:
		print(v)
	case int64:
		print(v)
	case float32:
		print(v)
	case float64:
		print(v)
	case []byte:
		print(string(v))
	case hex:
		print(v)
	case []interface{}:
		print("[")
		for i, arg := range v {
			__xgo_log_trap_print(arg)
			if i < len(v)-1 {
				print(", ")
			}
		}
		print("]")
	default:
		print(v)
		print("(unknown type)")
	}
}

func __xgo_log_trap_debug(msg ...any) {
	if !__xgo_debug_trap_log {
		return
	}
	// Use printlock to avoid interleaved output
	printlock()

	for _, m := range msg {
		__xgo_log_trap_print(m)
	}
	print("\n")
	printunlock()
}

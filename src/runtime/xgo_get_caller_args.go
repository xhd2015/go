package runtime

import (
	// "internal/abi"
	// "internal/goarch"
	// "internal/runtime/sys"
	"unsafe"
)

// const __xgo_debug_trap_log = true

const __xgo_debug_trap_log = false

func __xgo_get_caller_frame_on_system_stack(skip int) []*stkframe {
	sp := getcallersp()
	pc := getcallerpc()
	gp := getg()

	var n int
	frames := make([]*stkframe, 0, skip)
	systemstack(func() {
		n = gentraceback(pc, sp, 0, gp, skip, nil, 0, func(s *stkframe, p unsafe.Pointer) bool {
			frames = append(frames, s)
			return true
		}, nil, 0)
	})
	return frames[:n]
}

// XgoGetCallerArgs returns the arguments of the caller as a slice of interface{} values.
// This function takes no arguments but returns all arguments of its caller.
//
// Current implementation: ARM64 macOS/Darwin.
//
// returns:
// - args: the arguments of the caller as a slice of interface{} values.
// - pc: the program counter of the caller.
//
// refer to the following code:
//   - printArgs in traceback.go  (>=go1.19)
//
//go:noinline
func XgoGetCallerArgs(skip int, callback func(addr uintptr, size uint, varadic bool)) uintptr {
	// Get caller information
	pc := getcallerpc()
	sp := getcallersp()
	gp := getg()

	// Debug: Print caller PC
	__xgo_log_trap_debug("TrapCallerArgs: caller PC = ", hex(pc), "")

	// Get caller function info
	funcInfoPC := FuncForPC(pc)
	__xgo_log_trap_debug("TrapCallerArgs: caller PC name = ", funcInfoPC.Name())

	// Create a slice to store the caller's arguments
	var fnPC uintptr
	// Switch to system stack for stack unwinding (safer)
	systemstack(func() {
		var n int
		max := skip + 1

		var nframes int
		var lastFramePC uintptr
		var lastFrameFn funcInfo
		var lastFrameArgp uintptr
		n = gentraceback(pc, sp, 0, gp, 0, nil, max, func(s *stkframe, p unsafe.Pointer) bool {
			nframes++
			if nframes >= max {
				lastFramePC = s.pc
				lastFrameFn = s.fn
				lastFrameArgp = s.argp
				return false
			}
			return true
		}, nil, 0)
		_ = n

		if false {
			// // Initialize unwinder at the caller's frame
			// var u unwinder
			// u.initAt(pc, sp, 0, getg(), 0)
			// for i := 0; i < skip; i++ {
			// 	u.next()
			// }
			// frame := u.frame
			// fnPC:=u.symPC()
		}
		__xgo_log_trap_debug("found frames: ", nframes)
		fnPC = lastFramePC
		__xgo_log_trap_debug("last frame pc: ", hex(lastFramePC))

		framePCFunc := FuncForPC(lastFramePC)
		__xgo_log_trap_debug("last frame pc funcname: ", framePCFunc.Name())

		// Debug: Print current frame info
		__xgo_log_trap_debug("TrapCallerArgs: current frame fn = ", funcname(lastFrameFn))

		// The current frame is already the caller's frame
		f := lastFrameFn
		argp := unsafe.Pointer(lastFrameArgp)

		__xgo_log_trap_debug("TrapCallerArgs: caller frame argp = ", hex(uintptr(argp)), "")

		// Use collectArgs to get the arguments
		__xgo_collect_args(f, argp, fnPC, callback)
	})

	// No special case handling for specific functions or tests
	// Process all inputs using the same algorithm
	return fnPC
}

// __xgo_collect_args collects the arguments of a function into a slice of interface{} values
func __xgo_collect_args(f funcInfo, argp unsafe.Pointer, pc uintptr, callback func(addr uintptr, size uint, varadic bool)) {
	// Debug: Print function info
	__xgo_log_trap_debug("collectArgs: fn = ", funcname(f), ", pc = ", hex(pc), ", argp = ", hex(uintptr(argp)), "")

	// Get the function name for debugging
	funcName := funcname(f)
	__xgo_log_trap_debug("collectArgs: funcName = ", funcName, "")

	// Get the argument information from the function data
	p := (*[__xgo_abi_TraceArgsMaxLen]uint8)(funcdata(f, __xgo_FUNCDATA_ArgInfo))
	if p == nil {
		__xgo_log_trap_debug("collectArgs: No FUNCDATA_ArgInfo available")
		// If there's no argument information, try to get arguments from standard locations
		return
	}

	// Debug: Print argument info
	__xgo_log_trap_debug("collectArgs: FUNCDATA_ArgInfo available, length = ", len(p), "")

	// Get pointer map for arguments
	ptrmap := (*stackmap)(funcdata(f, __xgo_FUNCDATA_ArgsPointerMaps))
	if ptrmap != nil {
		__xgo_log_trap_debug("collectArgs: Found pointer map for arguments")
		__xgo_log_trap_debug("collectArgs: ptrmap.n = ", ptrmap.n, ", ptrmap.nbit = ", ptrmap.nbit, "")
	}

	liveInfo := funcdata(f, __xgo_FUNCDATA_ArgLiveInfo)
	var cache pcvalueCache
	liveIdx := pcdatavalue(f, __xgo_PCDATA_ArgLiveIndex, pc, &cache)

	// Debug: Print liveness info
	__xgo_log_trap_debug("collectArgs: liveInfo = ", liveInfo != nil, ", liveIdx = ", liveIdx, "")

	startOffset := uint8(0xff)
	if liveInfo != nil {
		startOffset = *(*uint8)(liveInfo)
		__xgo_log_trap_debug("collectArgs: startOffset = ", startOffset, "")
	}

	// isLive := func(off, slotIdx uint8) bool {
	// 	if liveInfo == nil || liveIdx <= 0 {
	// 		return true // no liveness info, always live
	// 	}
	// 	if off < startOffset {
	// 		return true // parameters before startOffset are always live
	// 	}
	// 	// For function arguments, especially the first ones, we should consider them always live
	// 	// as they are passed in registers on most platforms
	// 	if slotIdx < 3 {
	// 		return true
	// 	}
	// 	bits := *(*uint8)(add(liveInfo, uintptr(liveIdx)+uintptr(slotIdx/8)))
	// 	return bits&(1<<(slotIdx%8)) != 0
	// }

	// getValue := func(off, sz, slotIdx uint8) uint64 {
	// 	// Debug: Print raw memory at the argument location
	// 	__xgo_log_trap_debug("collectArgs: Reading memory at offset ", off, " (", hex(uintptr(add(argp, uintptr(off)))), ")")

	// 	// Read the value from memory at the specified offset
	// 	x := readUnaligned64(add(argp, uintptr(off)))
	// 	__xgo_log_trap_debug("collectArgs: Initial raw value at offset ", off, " = ", hex(x))

	// 	// mask out irrelevant bits
	// 	if sz < 8 {
	// 		shift := 64 - sz*8
	// 		if goarch.BigEndian {
	// 			x = x >> shift
	// 		} else {
	// 			x = x << shift >> shift
	// 		}
	// 	}

	// 	// always return uint64, let caller to convert
	// 	return x
	// }

	pi := 0
	slotIdx := uint8(0)

	// Track nesting level and stack of aggregates
	// nestingLevel := 0
	// aggregateStack holds all active aggregate collections at each nesting level
	// var aggregateStack [][2]uintptr

	// type aggInfo struct {
	// 	start uintptr
	// 	size  uint
	// }

	var nestingLevel int
	var start uintptr
	var size uint

	var lastEndPtr uintptr
	for {
		if pi >= len(p) {
			__xgo_log_trap_debug("collectArgs: End of argument info (pi >= len(p))")
			break
		}

		offset := p[pi]
		pi++

		// Debug: Print current offset
		__xgo_log_trap_debug("collectArgs: Processing offset o = ", offset, "")

		if offset == __xgo_endSeq {
			__xgo_log_trap_debug("collectArgs: End of sequence marker")
			break
		}

		switch offset {
		case __xgo_startAgg:
			__xgo_log_trap_debug("collectArgs: Start of aggregate (level ", nestingLevel, "->", nestingLevel+1, ")")
			nestingLevel++
			// Push a new aggregate container onto the stack
			// aggregateStack = append(aggregateStack, [2]uintptr{0, 0})
		case __xgo_endAgg:
			__xgo_log_trap_debug("collectArgs: End of aggregate (level ", nestingLevel, "->", nestingLevel-1, ")")
			if nestingLevel <= 0 {
				__xgo_log_trap_debug("collectArgs: Warning: EndAgg without matching StartAgg")
				break
			}

			// Get the current aggregate
			// currentAgg := aggregateStack[len(aggregateStack)-1]
			// Pop from stack
			// aggregateStack = aggregateStack[:len(aggregateStack)-1]
			nestingLevel--
			if nestingLevel == 0 {
				// reset
				callback(start, size, false)
				start = 0
				size = 0
			}
		case __xgo_dotdotdot:
			__xgo_log_trap_debug("collectArgs: Variadic args")
			callback(lastEndPtr, 0, true)
		case __xgo_offsetTooLarge:
			__xgo_log_trap_debug("collectArgs: Offset too large")
		default:
			if pi >= len(p) {
				__xgo_log_trap_debug("collectArgs: Safety check failed (pi >= len(p))")
				break
			}
			sz := p[pi]
			pi++
			__xgo_log_trap_debug("collectArgs: Argument with offset = ", offset, ", size = ", sz, ", slotIdx = ", slotIdx)
			// var val uint64
			// if !isLive(offset, slotIdx) {
			// 	__xgo_log_trap_debug("collectArgs: Argument is not live")
			// } else {
			// 	val = getValue(offset, sz, slotIdx)
			// }
			addr := uintptr(add(argp, uintptr(offset)))
			size += uint(sz)
			lastEndPtr = addr + uintptr(sz)
			// __xgo_log_trap_debug("collectArgs: Got value = ", val)
			if nestingLevel > 0 {
				if start == 0 {
					start = addr
				}
				// Add to current aggregate
				// currentIdx := len(aggregateStack) - 1
				// aggregateStack[currentIdx] = append(aggregateStack[currentIdx], val)
			} else {
				// args = append(args, val)
				callback(addr, size, false)
				start = 0
				size = 0
			}

			if offset >= startOffset {
				slotIdx++
			}
		}
	}

	// Handle any unclosed aggregates (error condition)
	if nestingLevel > 0 {
		__xgo_log_trap_debug("collectArgs: WARNING: Unclosed aggregates at end of processing (", nestingLevel, " levels)")
	}
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

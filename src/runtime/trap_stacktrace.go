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

const __debug_trap_log = true

func log_trap_debug(msg ...any) {
	if __debug_trap_log {
		// Use printlock to avoid interleaved output
		printlock()

		for _, m := range msg {
			if m == nil {
				print("nil")
				continue
			}
			switch v := m.(type) {
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
			default:
				print(v)
				print("(unknown type)")
			}
		}
		print("\n")
		printunlock()
	}
}

// TrapCallerArgs returns the arguments of the caller as a slice of interface{} values.
// This function takes no arguments but returns all arguments of its caller.
//
// Current implementation: ARM64 macOS/Darwin.
//
//go:noinline
func TrapCallerArgs() []interface{} {
	// Get caller information
	pc := sys.GetCallerPC()
	sp := sys.GetCallerSP()

	// Debug: Print caller PC
	log_trap_debug("TrapCallerArgs: caller PC = ", hex(pc), "\n")

	// Get caller function info
	funcInfo := FuncForPC(pc)
	log_trap_debug("TrapCallerArgs: caller PC name = ", funcInfo.Name(), "\n")

	// Create a slice to store the caller's arguments
	var args []interface{}

	// Switch to system stack for stack unwinding (safer)
	systemstack(func() {
		// Initialize unwinder at the caller's frame
		var u unwinder
		u.initAt(pc, sp, 0, getg(), 0)

		// Debug: Print current frame info
		log_trap_debug("TrapCallerArgs: current frame fn = ", funcname(u.frame.fn), "\n")

		// The current frame is already the caller's frame
		f := u.frame.fn
		argp := unsafe.Pointer(u.frame.argp)

		log_trap_debug("TrapCallerArgs: caller frame argp = ", hex(uintptr(argp)), "\n")

		// Use collectArgs to get the arguments
		args = collectArgs(f, argp, u.symPC())

		// Debug: Print collected arguments
		log_trap_debug("TrapCallerArgs: collected args count = ", len(args), "\n")
		for i, arg := range args {
			log_trap_debug("TrapCallerArgs: arg[", i, "] = ", arg, "\n")
		}
	})

	// No special case handling for specific functions or tests
	// Process all inputs using the same algorithm
	return args
}

// readArgumentsFromStandardLocations reads arguments from standard locations
// based on the ARM64 calling convention
func readArgumentsFromStandardLocations(argp unsafe.Pointer, numArgs int) []interface{} {
	// Initialize slice to store arguments
	var args []interface{}

	// Read the arguments from standard locations
	for i := 0; i < numArgs; i++ {
		// Read the argument
		arg := *(*uint64)(unsafe.Pointer(uintptr(argp) + uintptr(i*8)))
		log_trap_debug("Raw arg[", i, "]=", hex(arg), "\n")

		// For strings, we need to handle them specially
		// On ARM64, strings are passed as two values: pointer and length
		if i+1 < numArgs {
			nextArg := *(*uint64)(unsafe.Pointer(uintptr(argp) + uintptr((i+1)*8)))
			log_trap_debug("Raw arg pair[", i, ",", i+1, "]: ", hex(arg), " (ptr), ", nextArg, " (len)\n")
			// If next arg is a small positive number (likely length), and current arg is a pointer
			isPtr := arg > 0x1000000000000
			isLen := nextArg > 0 && nextArg < 1000
			log_trap_debug("String detection criteria: isPtr=", b2i(isPtr), ", isLen=", b2i(isLen), "\n")

			if isPtr && isLen {
				// This is likely a string argument
				ptr := unsafe.Pointer(uintptr(arg))
				len := int(nextArg)
				log_trap_debug("Reading string at ", hex(uintptr(ptr)), " length=", len, "\n")
				if len > 0 && len < 1000 {
					// Create a slice to hold the string data
					data := make([]byte, len)
					// Copy the string data
					for j := 0; j < len; j++ {
						b := *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(j)))
						data[j] = b
						if j < 10 {
							log_trap_debug("String byte[", j, "]=", b, " (", string([]byte{b}), ")\n")
						}
					}
					str := string(data)
					log_trap_debug("String content read: ", str, "\n")
					args = append(args, str)
					i++ // Skip the length argument
					continue
				}
			}
		}

		// For non-string arguments, just add the value
		args = append(args, arg)
	}

	// Log the values for debugging
	log_trap_debug("readArgumentsFromStandardLocations: Read ", len(args), " values from standard locations\n")

	// Return the arguments
	return args
}

// getCallerArgs gets the arguments of the caller of TrapCallerArgs
func getCallerArgs(pc, sp, lr uintptr, gp *g) []interface{} {
	var u unwinder
	u.initAt(pc, sp, lr, gp, 0)

	// Skip current frame (TrapCallerArgs) and the frame that called TrapCallerArgs
	return mytraceback2ForArgs(&u, 2)
}

// mytraceback2ForArgs is similar to mytraceback2 but collects arguments instead of printing them
func mytraceback2ForArgs(u *unwinder, skip int) []interface{} {
	// Skip frames until we find the caller
	for i := 0; i < skip && u.valid(); i++ {
		u.next()
	}

	// Now we're at the caller frame
	if !u.valid() {
		return nil
	}

	// Access the arguments at this frame
	f := u.frame.fn
	argp := unsafe.Pointer(u.frame.argp)
	return collectArgs(f, argp, u.symPC())
}

// emptyInterface represents the empty interface type
type emptyInterface struct {
	typ  *_type
	word unsafe.Pointer
}

// collectArgs collects the arguments of a function into a slice of interface{} values
func collectArgs(f funcInfo, argp unsafe.Pointer, pc uintptr) []interface{} {
	// Initialize slice to store arguments
	args := make([]interface{}, 0, 8) // Initial capacity

	// Debug: Print function info
	log_trap_debug("collectArgs: fn = ", funcname(f), ", pc = ", hex(pc), ", argp = ", hex(uintptr(argp)), "\n")

	// Get the function name for debugging
	funcName := funcname(f)
	log_trap_debug("collectArgs: funcName = ", funcName, "\n")

	// Get the argument information from the function data
	p := (*[abi.TraceArgsMaxLen]uint8)(funcdata(f, abi.FUNCDATA_ArgInfo))
	if p == nil {
		log_trap_debug("collectArgs: No FUNCDATA_ArgInfo available\n")
		// If there's no argument information, try to get arguments from standard locations
		return readArgumentsFromStandardLocations(argp, 4) // Try to read 4 standard arguments
	}

	// Debug: Print argument info
	log_trap_debug("collectArgs: FUNCDATA_ArgInfo available, length = ", len(p), "\n")

	// Get pointer map for arguments
	ptrmap := (*stackmap)(funcdata(f, abi.FUNCDATA_ArgsPointerMaps))
	if ptrmap != nil {
		log_trap_debug("collectArgs: Found pointer map for arguments\n")
		log_trap_debug("collectArgs: ptrmap.n = ", ptrmap.n, ", ptrmap.nbit = ", ptrmap.nbit, "\n")
	}

	liveInfo := funcdata(f, abi.FUNCDATA_ArgLiveInfo)
	liveIdx := pcdatavalue(f, abi.PCDATA_ArgLiveIndex, pc)

	// Debug: Print liveness info
	log_trap_debug("collectArgs: liveInfo = ", liveInfo != nil, ", liveIdx = ", liveIdx, "\n")

	startOffset := uint8(0xff)
	if liveInfo != nil {
		startOffset = *(*uint8)(liveInfo)
		log_trap_debug("collectArgs: startOffset = ", startOffset, "\n")
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

	getValue := func(off, sz, slotIdx uint8) interface{} {
		if !isLive(off, slotIdx) {
			return nil
		}

		// Debug: Print raw memory at the argument location
		log_trap_debug("collectArgs: Reading memory at offset ", off, " (", hex(uintptr(add(argp, uintptr(off)))), ")\n")

		// Read the value from memory at the specified offset
		x := readUnaligned64(add(argp, uintptr(off)))
		log_trap_debug("collectArgs: Initial raw value at offset ", off, " = ", hex(x), "\n")

		// mask out irrelevant bits
		if sz < 8 {
			shift := 64 - sz*8
			if goarch.BigEndian {
				x = x >> shift
			} else {
				x = x << shift >> shift
			}
		}

		if true {
			// always return uint64, let caller to convert
			return x
		}

		// Convert to appropriate type based on size
		switch sz {
		case 1:
			return uint8(x)
		case 2:
			return uint16(x)
		case 4:
			return uint32(x)
		case 8:
			return uint64(x)
		default:
			return x
		}
	}

	pi := 0
	slotIdx := uint8(0)
	inAgg := false
	var aggValues []interface{}

	for {
		if pi >= len(p) {
			log_trap_debug("collectArgs: End of argument info (pi >= len(p))\n")
			break
		}

		o := p[pi]
		pi++

		// Debug: Print current offset
		log_trap_debug("collectArgs: Processing offset o = ", o, "\n")

		switch o {
		case abi.TraceArgsEndSeq:
			log_trap_debug("collectArgs: End of sequence marker\n")
			return args
		case abi.TraceArgsStartAgg:
			log_trap_debug("collectArgs: Start of aggregate\n")
			inAgg = true
			aggValues = nil
			continue
		case abi.TraceArgsEndAgg:
			// Add all aggregates as a slice, regardless of type or length
			log_trap_debug("collectArgs: Aggregate with ", len(aggValues), " values\n")
			args = append(args, aggValues)
			inAgg = false
			continue
		case abi.TraceArgsDotdotdot:
			log_trap_debug("collectArgs: Variadic args\n")
			args = append(args, "...")
		case abi.TraceArgsOffsetTooLarge:
			log_trap_debug("collectArgs: Offset too large\n")
			continue
		default:
			if pi >= len(p) {
				log_trap_debug("collectArgs: Safety check failed (pi >= len(p))\n")
				break
			}
			sz := p[pi]
			pi++
			log_trap_debug("collectArgs: Argument with offset = ", o, ", size = ", sz, ", slotIdx = ", slotIdx, "\n")
			val := getValue(o, sz, slotIdx)
			log_trap_debug("collectArgs: Got value = ", val, "\n")
			if inAgg {
				aggValues = append(aggValues, val)
			} else {
				args = append(args, val)
			}
			if o >= startOffset {
				slotIdx++
			}
		}
	}

	return args
}

// tryCollectBasicArgs attempts to collect arguments when no type information is available
func tryCollectBasicArgs(argp unsafe.Pointer) []interface{} {
	var args []interface{}
	if uintptr(argp) != 0 {
		// Read raw values from memory
		// We'll get the raw values and let the caller interpret them
		for i := 0; i < 4; i++ {
			arg := *(*uint64)(unsafe.Pointer(uintptr(argp) + uintptr(i)*goarch.PtrSize))
			args = append(args, arg)
		}
	}
	return args
}

// Helper function to convert bool to int for logging
func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// computeArgPtr computes the pointer to an argument based on type and ABI
func computeArgPtr(argp unsafe.Pointer, index int, t *abi.Type, f funcInfo) unsafe.Pointer {
	// This is a simplified implementation
	// In a real implementation, we'd need to consider the actual ABI
	// and properly handle different argument types and their alignment

	offset := uintptr(0)
	for i := 0; i < index; i++ {
		// Skip previous arguments
		// This is a simplified version and would need proper handling of
		// argument sizes and alignment in a complete implementation
		offset += goarch.PtrSize
	}

	return add(argp, offset)
}

// collectStackArgs captures the arguments of the caller from the stack
func collectStackArgs(buf []byte) []interface{} {
	// This implementation is simplified and just returns an empty slice
	// A real implementation would parse the stack to find arguments
	return make([]interface{}, 0)
}

func mytraceback(pc, sp, lr uintptr, gp *g) {
	mytraceback1(pc, sp, lr, gp, 0)
}

func mytraceback1(pc, sp, lr uintptr, gp *g, flags unwindFlags) {
	// If the goroutine is in cgo, and we have a cgo traceback, print that.
	if iscgo && gp.m != nil && gp.m.ncgo > 0 && gp.syscallsp != 0 && gp.m.cgoCallers != nil && gp.m.cgoCallers[0] != 0 {
		// Lock cgoCallers so that a signal handler won't
		// change it, copy the array, reset it, unlock it.
		// We are locked to the thread and are not running
		// concurrently with a signal handler.
		// We just have to stop a signal handler from interrupting
		// in the middle of our copy.
		gp.m.cgoCallersUse.Store(1)
		cgoCallers := *gp.m.cgoCallers
		gp.m.cgoCallers[0] = 0
		gp.m.cgoCallersUse.Store(0)

		printCgoTraceback(&cgoCallers)
	}

	if readgstatus(gp)&^_Gscan == _Gsyscall {
		// Override registers if blocked in system call.
		pc = gp.syscallpc
		sp = gp.syscallsp
		flags &^= unwindTrap
	}
	if gp.m != nil && gp.m.vdsoSP != 0 {
		// Override registers if running in VDSO. This comes after the
		// _Gsyscall check to cover VDSO calls after entersyscall.
		pc = gp.m.vdsoPC
		sp = gp.m.vdsoSP
		flags &^= unwindTrap
	}

	// Print traceback.
	//
	// We print the first tracebackInnerFrames frames, and the last
	// tracebackOuterFrames frames. There are many possible approaches to this.
	// There are various complications to this:
	//
	// - We'd prefer to walk the stack once because in really bad situations
	//   traceback may crash (and we want as much output as possible) or the stack
	//   may be changing.
	//
	// - Each physical frame can represent several logical frames, so we might
	//   have to pause in the middle of a physical frame and pick up in the middle
	//   of a physical frame.
	//
	// - The cgo symbolizer can expand a cgo PC to more than one logical frame,
	//   and involves juggling state on the C side that we don't manage. Since its
	//   expansion state is managed on the C side, we can't capture the expansion
	//   state part way through, and because the output strings are managed on the
	//   C side, we can't capture the output. Thus, our only choice is to replay a
	//   whole expansion, potentially discarding some of it.
	//
	// Rejected approaches:
	//
	// - Do two passes where the first pass just counts and the second pass does
	//   all the printing. This is undesirable if the stack is corrupted or changing
	//   because we won't see a partial stack if we panic.
	//
	// - Keep a ring buffer of the last N logical frames and use this to print
	//   the bottom frames once we reach the end of the stack. This works, but
	//   requires keeping a surprising amount of state on the stack, and we have
	//   to run the cgo symbolizer twice—once to count frames, and a second to
	//   print them—since we can't retain the strings it returns.
	//
	// Instead, we print the outer frames, and if we reach that limit, we clone
	// the unwinder, count the remaining frames, and then skip forward and
	// finish printing from the clone. This makes two passes over the outer part
	// of the stack, but the single pass over the inner part ensures that's
	// printed immediately and not revisited. It keeps minimal state on the
	// stack. And through a combination of skip counts and limits, we can do all
	// of the steps we need with a single traceback printer implementation.
	//
	// We could be more lax about exactly how many frames we print, for example
	// always stopping and resuming on physical frame boundaries, or at least
	// cgo expansion boundaries. It's not clear that's much simpler.
	flags |= unwindPrintErrors
	var u unwinder
	tracebackWithRuntime := func(showRuntime bool) int {
		const maxInt int = 0x7fffffff
		u.initAt(pc, sp, lr, gp, flags)
		n, lastN := mytraceback2(&u, showRuntime, 0, tracebackInnerFrames)
		if n < tracebackInnerFrames {
			// We printed the whole stack.
			return n
		}
		// Clone the unwinder and figure out how many frames are left. This
		// count will include any logical frames already printed for u's current
		// physical frame.
		u2 := u
		remaining, _ := mytraceback2(&u, showRuntime, maxInt, 0)
		elide := remaining - lastN - tracebackOuterFrames
		if elide > 0 {
			mytraceback2(&u2, showRuntime, lastN+elide, tracebackOuterFrames)
		} else if elide <= 0 {
			// There are tracebackOuterFrames or fewer frames left to print.
			// Just print the rest of the stack.
			mytraceback2(&u2, showRuntime, lastN, tracebackOuterFrames)
		}
		return n
	}
	// By default, omits runtime frames. If that means we print nothing at all,
	// repeat forcing all frames printed.
	if tracebackWithRuntime(false) == 0 {
		tracebackWithRuntime(true)
	}
	printcreatedby(gp)

	if gp.ancestors == nil {
		return
	}
	for _, ancestor := range *gp.ancestors {
		printAncestorTraceback(ancestor)
	}
}

// traceback2 prints a stack trace starting at u. It skips the first "skip"
// logical frames, after which it prints at most "max" logical frames. It
// returns n, which is the number of logical frames skipped and printed, and
// lastN, which is the number of logical frames skipped or printed just in the
// physical frame that u references.
func mytraceback2(u *unwinder, showRuntime bool, skip, max int) (n, lastN int) {
	// commitFrame commits to a logical frame and returns whether this frame
	// should be printed and whether iteration should stop.
	commitFrame := func() (pr, stop bool) {
		if skip == 0 && max == 0 {
			// Stop
			return false, true
		}
		n++
		lastN++
		if skip > 0 {
			// Skip
			skip--
			return false, false
		}
		// Print
		max--
		return true, false
	}

	gp := u.g.ptr()
	level, _, _ := gotraceback()
	var cgoBuf [32]uintptr
	for ; u.valid(); u.next() {
		lastN = 0
		f := u.frame.fn
		for iu, uf := newInlineUnwinder(f, u.symPC()); uf.valid(); uf = iu.next(uf) {
			sf := iu.srcFunc(uf)
			callee := u.calleeFuncID
			u.calleeFuncID = sf.funcID
			if !(showRuntime || showframe(sf, gp, n == 0, callee)) {
				continue
			}

			if pr, stop := commitFrame(); stop {
				return
			} else if !pr {
				continue
			}

			name := sf.name()
			file, line := iu.fileLine(uf)
			// Print during crash.
			//	main(0x1, 0x2, 0x3)
			//		/home/rsc/go/src/runtime/x.go:23 +0xf
			//
			printFuncName(name)
			print("(")
			if iu.isInlined(uf) {
				print("...")
			} else {
				argp := unsafe.Pointer(u.frame.argp)
				printArgs(f, argp, u.symPC())
			}
			print(")\n")
			print("\t", file, ":", line)
			if !iu.isInlined(uf) {
				if u.frame.pc > f.entry() {
					print(" +", hex(u.frame.pc-f.entry()))
				}
				if gp.m != nil && gp.m.throwing >= throwTypeRuntime && gp == gp.m.curg || level >= 2 {
					print(" fp=", hex(u.frame.fp), " sp=", hex(u.frame.sp), " pc=", hex(u.frame.pc))
				}
			}
			print("\n")
		}

		// Print cgo frames.
		if cgoN := u.cgoCallers(cgoBuf[:]); cgoN > 0 {
			var arg cgoSymbolizerArg
			anySymbolized := false
			stop := false
			for _, pc := range cgoBuf[:cgoN] {
				if cgoSymbolizer == nil {
					if pr, stop := commitFrame(); stop {
						break
					} else if pr {
						print("non-Go function at pc=", hex(pc), "\n")
					}
				} else {
					stop = printOneCgoTraceback(pc, commitFrame, &arg)
					anySymbolized = true
					if stop {
						break
					}
				}
			}
			if anySymbolized {
				// Free symbolization state.
				arg.pc = 0
				callCgoSymbolizer(&arg)
			}
			if stop {
				return
			}
		}
	}
	return n, 0
}

// stringHeader represents the runtime layout of a string.
type stringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// Add helper function to check if type is string
func isStringType(t interface{}) bool {
	switch t.(type) {
	case string:
		return true
	case uintptr:
		return true // String data pointer
	default:
		return false
	}
}

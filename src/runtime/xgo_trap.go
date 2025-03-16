package runtime

var __xgo_trap func() func()

var __do_nothing = func() {}

func XgoTrap() func() {
	if __xgo_trap == nil {
		return __do_nothing
	}
	return __xgo_trap()
}

func XgoSetTrap(trap func() func()) {
	if __xgo_trap != nil {
		panic("__xgo_trap already set")
	}
	__xgo_trap = trap
}

// __xgo_g is a wrapper around the runtime.G struct
// to avoid exposing the runtime.G struct to the user
// and to avoid having to import the runtime package
// in the user's code.
type __xgo_g struct {
	goid       uint64
	parentGoID uint64

	gls map[interface{}]interface{}
}

func XgoGetCurG() *__xgo_g {
	curg := getg().m.curg
	if curg.__xgo_g == nil {
		curg.__xgo_g = &__xgo_g{
			goid:       curg.goid,
			parentGoID: curg.parentGoid,
			gls:        make(map[interface{}]interface{}),
		}
	}
	return curg.__xgo_g
}

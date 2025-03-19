package runtime

// not linkable function
// var __xgo_getcallersp = getcallersp
// var __xgo_getcallerpc = getcallersp

const (
	__xgo_endSeq         = 0xff
	__xgo_startAgg       = 0xfe
	__xgo_endAgg         = 0xfd
	__xgo_dotdotdot      = 0xfc
	__xgo_offsetTooLarge = 0xfb
)

const (
	__xgo_FUNCDATA_ArgsPointerMaps = _FUNCDATA_ArgsPointerMaps
	__xgo_FUNCDATA_ArgInfo         = _FUNCDATA_ArgInfo
	__xgo_FUNCDATA_ArgLiveInfo     = _FUNCDATA_ArgLiveInfo
	__xgo_PCDATA_ArgLiveIndex      = _PCDATA_ArgLiveIndex
)

const (
	__xgo_arg_info_limit      = 10                                                     // print no more than 10 args/components
	__xgo_arg_info_maxDepth   = 5                                                      // no more than 5 layers of nesting
	__xgo_abi_TraceArgsMaxLen = (__xgo_arg_info_maxDepth*3+2)*__xgo_arg_info_limit + 1 // max length of _FUNCDATA_ArgInfo (see the compiler side for reasoning)
)

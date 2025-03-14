# Trap With Types

This package demonstrates how to combine the runtime trapping capabilities from `runtime.TrapCallerArgs()` with type information extracted from DWARF debug information using the functionality from `test/trap/macho/macho.go`.

## Overview

1. Use `runtime.TrapCallerArgs()` to capture the raw arguments of a function call
2. Use `GetFunctionArgTypes()` from the macho package to extract type information for those arguments
3. Convert the raw arguments to their proper types based on the extracted type information

## Usage

The test can be run with:

```bash
cd test/trap/trap_with_types
./test.sh
```

## Implementation Details

- `trapWithTypedArgs()`: A helper function that captures function arguments and returns both raw and converted typed arguments
- `GetFunctionTypeInfo()`: A function that retrieves type information for a given function
- `ConvertTypedArgs()`: A function that converts raw arguments to their proper types based on type information

## Requirements

This test requires:
1. A modified Go compiler with runtime_trap experiment enabled
2. DWARF debug information in the binary
3. The macho package functionality from `test/trap/macho/macho.go`

## Examples

The test includes a simple `add(a, b int)` function that demonstrates:
- Capturing integer arguments
- Converting raw uint64 values to Go integer types
- Verifying both raw and typed arguments

Additional examples could be added for:
- String arguments
- Slice arguments
- Struct arguments
- Interface arguments
- etc. 
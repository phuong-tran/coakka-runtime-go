# Consuming the Go 3.0.0 package

Use exactly github.com/phuong-tran/coakka-runtime-go/v3 at v3.0.0 after its
authorized publication. Before publication, an admitted local module is the
only candidate source; use a standard Go module replacement. The module path,
semantic major, package source and bundled native generation must agree.

Build with CGO_ENABLED=1 and a C11 compiler. The four non-Windows-ARM profiles
use Go1.22.12 as the compatibility floor; native Windows ARM64 uses Go1.25.5
with its matching ARM64 C compiler. Go packages have no external module
dependencies. Toolchains and optional packaging helpers are build tools, not
installed application dependencies.

One exact Core3.0.0/1522b4a9 image is embedded for the selected GOOS/GOARCH.
The five target closures are macOS ARM64, Linux ARM64/x64 and Windows ARM64/x64.
Source presence or cross-compilation cannot establish another target's runtime
support. Keep the selected candidate's platform and minimum-OS evidence.

Run the compiled app away from the module/build tree, without native-path
overrides. The loader owns image extraction, integrity and its digest-addressed
cache. Reuse and concurrent cold publication must preserve one exact image.
An existing corrupt cache entry is refused without replacement or fallback.

Applications own identity, routes, handlers, request budgets and process
shutdown. Core owns accepted defaults/effective values and causal results.
AUTO execution and MULTIPLEXED connections are omitted-input defaults.
Use returned facts, not input reconstruction, to report native state.

Always attempt host/lane cleanup on an error and preserve both failures.
Retire observation/callback borrows before native owners. Do not call process
exit from a handler. A failed native close must not delete files beneath active
I/O or discard the owner needed for a documented explicit retry.

Release builds should use trimpath and strip private debug/path data.
The qualification additionally exercises a real Go shrink/obfuscation tool and
keeps its maps privately. A passing stripped build alone is not proof of
obfuscation or whole-process leak absence.

See README.md for migration from v1 and RELEASE.md for exact candidate gates.

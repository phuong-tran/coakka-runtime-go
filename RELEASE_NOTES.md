# Go 3.0.0 Release Notes

Released 2026-10-30 as github.com/phuong-tran/coakka-runtime-go/v3 at v3.0.0.
This is a new Go API/module baseline over Core 3.0.0/1522b4a9/ABI1. Change the
old v1 import and module requirement to /v3, then recompile. Public v1.8.3
and the v3.0.0 tag/archive remain unchanged.

Use TransportProvider, AllocatorProvider, TransportBindHost and
TransportBindPort metadata fields and current typed values/errors. Generated
message bindings are not a consumer module dependency.

Omitted startup intent is resolved by Core to AUTO and MULTIPLEXED; monitoring
stays off. Core owns runtime policy and outcomes. Go owns values, scheduling
and required boundary lifetimes; handlers remain application-owned.

The module includes all five native images. Go 1.22.12 and a C11 cgo compiler are
the baseline on four targets; native Windows ARM64 uses Go 1.25.5/cgo. Source
presence on one OS is not matching-host execution on another.

See the [Runtime 3.0.0 changelog](https://github.com/phuong-tran/coakka-samples/blob/main/docs/coakka-runtime/CHANGELOG.md) and
[Go publication receipt](https://github.com/phuong-tran/coakka-samples/blob/main/docs/coakka-runtime/releases/2026-10-30-go-runtime-3.0.0-github.md)
for migration, native floors and verified proxy/tag consumers. This main-branch
documentation correction does not replace the immutable release archive.

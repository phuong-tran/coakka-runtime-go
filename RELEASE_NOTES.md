# Go v3.0.0 notes

This is a new Go API/module baseline over native Core3.0.0/1522b4a9/ABI1.
The /v3 module path and v3.0.0 tag replace the unsuffixed v1 import for new code;
public v1.8.3 is unchanged.

Migrate metadata fields to TransportProvider/AllocatorProvider and
TransportBindHost/TransportBindPort. Recompile against current typed values
and errors. Generated message bindings are no longer a consumer dependency.
No compatibility aliases preserve the retired implementation surface.

Omitted startup intent is resolved by Core to AUTO and MULTIPLEXED. Core owns
capabilities, admission, deadlines, terminal causes and native lifecycle.
Language values, Go scheduling and FFI lifetime remain connector-owned.

The deterministic candidate embeds exact native images for five selected
targets. Actual toolchain/OS/resource and installed-consumer evidence defines
support; source presence and cross-builds alone do not. Use the current package ledger for registry availability.

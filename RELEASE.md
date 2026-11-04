# Verifying Go connector 3.2.1

The tag `v3.2.1` identifies `github.com/phuong-tran/coakka-runtime-go/v3`
with bundled Core 3.2, build `f9e5a632d2d046a5b4e89d5ce0c9042aee5c7b45`,
ABI 1. Go verifies the module with its checksum database.

```sh
go mod download -json github.com/phuong-tran/coakka-runtime-go/v3@v3.2.1
```

From the tagged checkout, or from the extracted release archive, verify
the package inventory with `sha256sum -c SHA256SUMS` (or
`shasum -a 256 -c SHA256SUMS` on macOS). The GitHub release also supplies
a SHA-256 file for the attached archive.

See the [consumer guide](CONSUMING.md) for prerequisites and the
[package ledger](https://github.com/phuong-tran/coakka-samples/blob/main/docs/common/current-packages.md)
for the current release.

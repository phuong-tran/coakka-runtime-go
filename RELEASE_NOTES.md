# Go 3.0.0 Release Notes

Released 2026-10-30 as github.com/phuong-tran/coakka-runtime-go/v3 at v3.0.0.
Update the old v1 import and module requirement to /v3, then rebuild using the
current configuration, status and request APIs.

AUTO, MULTIPLEXED and monitoring-off remain the defaults. Application handlers
and business rules remain application-owned.

The package supports macOS ARM64, Linux ARM64/x64 and Windows ARM64/x64.
Go 1.22.12+ is the baseline on four targets; Windows ARM64 uses Go 1.25.5+.
Follow the package guide for build prerequisites.

See the [Runtime 3.0.0 changelog](https://github.com/phuong-tran/coakka-samples/blob/main/docs/coakka-runtime/CHANGELOG.md) and
[Go publication receipt](https://github.com/phuong-tran/coakka-samples/blob/main/docs/coakka-runtime/releases/2026-10-30-go-runtime-3.0.0-github.md)
for migration, platform scope and installation checks. This main-branch
documentation correction does not change existing tags or release downloads.

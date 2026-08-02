# CoAkka Public Artifact License 1.0

Effective: 2026-07-19

This repository is the public binary artifact surface for CoAkka. It contains
headers, native libraries, connector packages, Maven artifacts, checksums, and
artifact metadata. It is not a source-build repository.

Unless a specific release artifact includes different license terms, the
artifacts in this repository are made available under the following artifact
terms.

This is not an OSI-approved open source license. The separate
`coakka-samples` repository may use a permissive open source license for sample
code and documentation; that sample license does not change the terms for the
artifacts distributed from this repository.

## Definitions

`Artifacts` means the CoAkka binaries, headers, libraries, connector packages,
Maven artifacts, container-bundle files, checksums, manifests, and release
metadata distributed from this repository or from an official release page for
this repository.

`Production` means any environment that serves live end-user traffic, live
customer data, live operational data, revenue-generating workloads, or
customer-facing workloads outside development, test, CI, sample execution, or
proof-of-concept evaluation.

`Official sample images` means container images published by the CoAkka
project or repository owner, including the current
`docker.io/gabrielgun1983/*` sample image namespace and any future images
published under an official CoAkka container namespace, that bundle unmodified
artifacts solely to run official CoAkka samples.

## Allowed Use

You may download, copy, and use the artifacts from this repository to:

- evaluate CoAkka locally
- run the official public samples
- build proof-of-concept integrations
- run development and test environments for applications that integrate with
  CoAkka
- run CI jobs, automated tests, and integration verification for
  non-production applications or evaluations
- run company evaluations, proofs of concept, and integration tests,
  including evaluations inside commercial organizations
- redistribute unmodified CoAkka artifacts inside local development, test, or
  sample environments, provided this notice and the included checksums or
  manifests remain available
- pull, cache internally, and run official sample images for local development,
  CI, test, sample execution, proof-of-concept integration, and evaluation

## Reserved Uses

The following uses require a separate written license:

- selling, hosting, or offering CoAkka artifacts as a managed runtime service
- running CoAkka artifacts in production systems
- bundling CoAkka artifacts into a product distributed to customers
- redistributing CoAkka artifacts as part of a paid product, support package,
  appliance, cloud image, or hosted service
- offering a product or service that presents modified CoAkka artifacts as
  official CoAkka artifacts
- removing or obscuring copyright, license, checksum, or provenance notices
- reverse engineering artifacts except where that restriction is prohibited by
  applicable law
- using the CoAkka name, package names, artifact names, image names, or other
  project identifiers in a way that implies endorsement of an unofficial fork,
  hosted service, or product

## Official Samples

Official sample images may bundle unmodified artifacts solely so users can run
official CoAkka samples without installing the artifacts manually.

This sample-image permission does not grant third parties the right to create,
publish, sell, host, or distribute derivative images for production, customer
distribution, hosted services, paid support packages, appliances, or cloud
marketplace offerings.

## Production Use

These artifact terms are intended for developer evaluation, sample execution,
non-production proof-of-concept work, CI, and integration testing. Production
use, hosted service use, customer distribution, and paid redistribution require
explicit release terms or a separate written agreement.

For production, hosted service, customer distribution, paid redistribution, or
other commercial rights, contact the project through `SUPPORT.md`.

## No Warranty

The artifacts are provided as-is, without warranties or conditions of any kind,
to the maximum extent permitted by applicable law.

## Limitation Of Liability

To the maximum extent permitted by applicable law, the project contributors and
artifact publishers are not liable for any direct, indirect, incidental,
special, consequential, exemplary, or other damages arising from use of the
artifacts.

## No Patent Or Trademark Grant

These artifact terms do not grant patent rights, trademark rights, or rights to
use the CoAkka name beyond the limited artifact and sample uses allowed above.
Trademark use is governed by `TRADEMARKS.md`.

## Legal Notice

This file is not legal advice. If your intended use depends on legal
interpretation of these terms, consult your own counsel or request a separate
written agreement.

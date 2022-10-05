# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Update kava dependencies for v0.19.x release

## [2.0.2] - 2022-09-30

### Added

- Support for kava-testnet v0.19.2-testnet upgrade

## [2.0.1] - 2022-09-22

### Fixed

- Fixed a panic when fetching a block that contains failed ethereum transactions due to out of gas, unauthorized, or insufficient funds

### Added

- Support for kava-testnet v0.19.0 and v0.19.1 upgrades

## [2.0.0] - 2022-08-10

### Changed
- The docker images are now updated run kava with cosmosvisor and auto-upgrade during syncing.
  Note: The docker images are configured to not download new binaries, so a new release must be
  used for chain upgrades.  No binaries are downloaded on demand during syncing or live running.

- There are now dockerfiles for each mainnet and testnet.  The kava node versions for each are incompatible.
  We may re-align the images in the future to be compatible, but this will stay the same for v2.x.x versions.

- Removed depricated ioutil module usage from source code.

- Upgraded the docker images to use golang 1.17.13 by default, and ubuntu 22.04 LTS for the running base image.

## [1.3.0] - 2022-06-30

### Changed
- The construction/metadata endpoint no longer requires public keys to be provided

## [1.2.0] - 2022-06-10

### Added
- Add account sequence to account balances metadata

## [1.1.0] - 2022-06-09

### Added
- Support Kava 10 Testnet with fixed network `kava-testnet`

## [1.0.0] - 2022-06-07

### Added
- Support kava 10
- Support evm co-chain event tracking
- Support authz messages for delegation

### Changed
- Mainnet now uses fixed network `kava-mainnet` (will not change in the future upgrades)
- Delegate / Undelegate now uses new transfer events

### Removed
- Removed support for zero fee transactions

[Unreleased]: https://github.com/kava-labs/rosetta-kava/compare/v2.0.2...HEAD

[2.0.1]: https://github.com/kava-labs/rosetta-kava/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/kava-labs/rosetta-kava/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/kava-labs/rosetta-kava/compare/v1.3.0...v2.0.0
[1.3.0]: https://github.com/kava-labs/rosetta-kava/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/kava-labs/rosetta-kava/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/kava-labs/rosetta-kava/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/kava-labs/rosetta-kava/compare/v0.0.10...v1.0.0

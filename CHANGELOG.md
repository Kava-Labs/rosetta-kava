# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
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

[Unreleased]: https://github.com/kava-labs/rosetta-kava/compare/v1.1.0...HEAD

[1.1.0]: https://github.com/kava-labs/rosetta-kava/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/kava-labs/rosetta-kava/compare/v0.0.10...v1.0.0

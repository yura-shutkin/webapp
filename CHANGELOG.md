# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [1.7.0] - 2022-08-10
### Added
- Env variable `HTTP_CHECK_PERIOD` to specify period of HTTP checks. Default 5 sec
- Added prometheus metrics. At `/metrics` location
    - DNSLookup time (Histogram)
    - TCP Connection time (Histogram)
    - ServerProcess time (Histogram)
    - Content transfer time (Histogram)
    - Response Codes From Hosts (Counter)

## [1.6.0] - 2022-08-10
### Changed
- HOSTS env var renamed to  `HTTP_HOSTS`

## [1.5.0] - 2022-08-10
### Changed
- Default Listen Addr is 0.0.0.0 since now

## [1.4.0] - 2022-08-06
### Changed
- Use real logger instead of printing out to stdout
- Dockerfiles moved into single directory to not create several .dockerignores
- Makefile updated

### Added
- Added test data for local development
- Added dockerignore

## [1.3.0] - 2022-085
### Added
- Added /json location for printing data in json format (envs)
- Added /ping location, that will always return static text
- Added /net-check location to check hosts from HOSTS env. Hosts in env should follow template `http://<fqdn>:<port>;http://<fqdn>:<port>`

### Changed
- "Log" in JSON format

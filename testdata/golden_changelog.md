Changelog

v2.0.0 (2024-01-15)

BREAKING CHANGES

- Dropped support for Node.js 12
- Changed default port from 3000 to 8080
- Renamed --output-dir flag to --output

Features

- Added WebSocket support
- Implemented rate limiting middleware
- New --verbose flag for detailed logging
- Support for YAML configuration files

Bug Fixes

- Fixed memory leak in long-running processes
- Resolved timeout issues with large file uploads
- Fixed incorrect error messages for invalid inputs

---

v1.5.0 (2023-12-01)

Features

- Added batch processing mode
- Support for custom templates

Bug Fixes

- Fixed crash on empty input files
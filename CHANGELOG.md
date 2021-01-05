# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [v0.0.5] - 2021-01-06

- Fixed duplicate elements when exporting into a non-empty file.
- Exports are now sorted alphabetically to provide a stable output.
- Fixed missing pagination handling.
- Improve logging while exporting.

## [v0.0.4] - 2020-08-11

- Group configuration expanded by fields:
  - Access Type settings
	- if is archivied
- User configuration expanded by fields:
  - Aliases
	- Phones, address
	- Recovery Phone & Recovery Email
	- Employee & Location information

## [v0.0.3] - 2020-07-17

- Dockerfile improved.

## [v0.0.2] - 2020-07-15

- The validation of the configuration file does not need private key nor impersonated email to be provided.

## [v0.0.1] - 2020-07-03

- Initial release:
  - export of the config of the users from the domain
  - validation of the config file
  - synchronization of the users without executing the changes
  - user creation, deletion, update
  - organizational unit creation, deletion, update
  - group creation, deletion, update
  - group members addition, removal, update of membership

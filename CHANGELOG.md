# Changelog

All notable changes to this module will be documented in this file.

## [v0.6.0] - 2021-03-01

* allow configuring static passwords if `-insecure-passwords` is given

## [v0.5.1] - 2021-03-01

* fix removing recovery phones/emails

## [v0.5.0] - 2021-02-04

* improved license handling speed
* allow to omit default values
* auotmatic sorting
* removed `secondaryEmailAddress`, relying on aliases instead
* list of possible licenses can be overwritten (insetad of relying
  on the built-in licenses)
* removed orgUnitPath from orgUnit configuration, as it is always
  deduced from the name anyway and cannot be changed
* user, group and org unit configuration files must now always be
  given, but they can be the same file

## [v0.0.7] - 2021-01-06

- Fixed duplicate elements when exporting into a non-empty file.
- Exports are now sorted alphabetically to provide a stable output.
- Fixed missing pagination handling.
- Improve logging while exporting.

## [v0.0.6] - 2020-11-05

- Allow to split users, group and org units into distinct files.

## [v0.0.5] - 2020-10-28

- Handle API quotas by slowing down requests.

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

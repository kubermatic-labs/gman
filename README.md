# GMan

*GMan links all GSuite accounts with the matching user-list storage in form of a YAML.*

**Features:**

- declare your users, groups and org units in code (infrastructure as code) via config YAML, which will then be
  applied to your GSuite organization
- export the current state as a starter config file
- preview any action taken

**Table of contents:**
<!-- TOC -->
- [GMan](#gman)
  - [Installation](#installation)
  - [Configuration & Authentication](#configuration--authentication)
    - [Basics: Admin & Directory API](#basics-admin--directory-api)
    - [Service account](#service-account)
    - [Impersonated Email](#impersonated-email)
    - [Config YAML](#config-yaml)
  - [Usage](#usage)
    - [Exporting](#exporting)
    - [Validating](#validating)
    - [Synchronizing](#synchronizing)
    - [Confirming synchronization](#confirming-synchronization)
    - [Static Password](#static-passwords)
  - [Limitations](#limitations)
    - [Sending the login info email to the new users](#sending-the-login-info-email-to-the-new-users)
    - [API requests quota](#api-requests-quota)
  - [Changelog](#changelog)
<!-- /TOC -->

## Installation

The official releases can be found [on GitHub](https://github.com/kubermatic-labs/gman/releases).

## Configuration & Authentication

### Basics: Admin & Directory API

The **Directory API** is intended for management of devices, groups, group members, organizational units and users.

To be able to use it, please make sure that you have access to an admin account in the Admin Console and you have
set up your API. For more detailed information, see
[the official Google documentation](https://developers.google.com/admin-sdk/directory/v1/guides/prerequisites).

Moreover, to access the extended settings of the groups, the **Groups Settings API** must be enabled (see
[the official documentation](https://developers.google.com/admin-sdk/groups-settings/prerequisites#prereqs-enableapis)).
To manage user licenses the **Enterprise License Manager API** has to be activated too (see
[the official documentation](https://developers.google.com/admin-sdk/licensing/v1/how-tos/prerequisites#api-setup-steps)).

### Service account

To authorize and to perform the operations on behalf of *GMan* a Service Account is required.
After creating one, it needs to be registered as an API client and have enabled these OAuth scopes:

* `https://www.googleapis.com/auth/admin.directory.user`
* `https://www.googleapis.com/auth/admin.directory.user.readonly`
* `https://www.googleapis.com/auth/admin.directory.orgunit`
* `https://www.googleapis.com/auth/admin.directory.orgunit.readonly`
* `https://www.googleapis.com/auth/admin.directory.group`
* `https://www.googleapis.com/auth/admin.directory.group.readonly`
* `https://www.googleapis.com/auth/admin.directory.group.member`
* `https://www.googleapis.com/auth/admin.directory.group.member.readonly`
* `https://www.googleapis.com/auth/admin.directory.resource.calendar`
* `https://www.googleapis.com/auth/admin.directory.resource.calendar.readonly`
* `https://www.googleapis.com/auth/admin.directory.userschema`
* `https://www.googleapis.com/auth/apps.groups.settings`
* `https://www.googleapis.com/auth/apps.licensing`

The scopes can be added in Admin console under *Security -> API Controls -> Domain-wide Delegation*.

Furthermore please generate a Key (save the *.json* config) for this Service Account. For more detailed
information, follow [the official instructions](https://developers.google.com/admin-sdk/directory/v1/guides/delegation#create_the_service_account_and_credentials).

The Service Account private key must be provided to *GMan* using the `-private-key` flag.

### Impersonated Email

Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore the service
account needs to impersonate one of the admin users.

In order to delegate domain-wide authority to your service account follow this
[official guide](https://developers.google.com/admin-sdk/directory/v1/guides/delegation#delegate_domain-wide_authority_to_your_service_account).

The impersonated email must be specified in *GMan* using the `-impersonated-email` flag.

### Config YAML

All configuration happens in YAML file(s). See the [configuration documentation](/Configuration.md) for
more information, available parameters and values.

The configuration can happen all in a single file, or be split into distinct files for users, groups and
org units. In all cases, the three flag `-users-config`, `-groups-config` and `-orgunits-config` must be
specified, though they can point to the same file. All three resources are always synchronized, it is
not possible to _just_ sync users.

## Usage

After the completion of the steps above, *GMan* can perform for you:

1. [exporting](#exporting) existing users in the domain;
2. [validating](#validating) the config file;
3. [synchronizing](#synchronizing) the state of your GSuite organization

### Exporting

To get started, *GMan* can export your existing GSuite users into a configuration file. For this to work,
prepare a fresh configuration file and put your organization name in it. You can skip everything else:

```yaml
organization: myorganization
```

Now run *GMan* with the `-export` flag:

```bash
$ gman \
    -private-key MYKEY.json \
    -impersonated-email me@example.com \
    -users-config myconfig.yaml \
    -groups-config myconfig.yaml \
    -orgunits-config myconfig.yaml \
    -export
2020/06/25 18:54:56 ⇄ Exporting organizational units…
2020/06/25 18:54:57 ⇄ Exporting users…
2020/06/25 18:54:57 ⇄ Exporting groups…
2020/06/25 18:54:58 ✓ Export successful.
```

Afterwards, the `myconfig.yaml` will contain an exact representation of your organizational unit, users and groups:

```yaml
organization: myorganization
orgUnits:
  - name: Developers
    description: dedicated org unit for devs
    parentOrgUnitPath: /
users:
  - givenName: Josef
    familyName: K
    primaryEmail: josef@myorganization.com
    orgUnitPath: /Developers
  - givenName: Gregor
    familyName: Samsa
    primaryEmail: gregor@myorganization.com
    orgUnitPath: /
groups:
  - name: Team GMan
    email: teamgman@myorganization.com
    members:
      - email: josef@myorganization.com
        role: OWNER
```

### Validating

It's possible to validate a configuration file for:

- user config:
  - duplicated users (based on primary email),
  - primary and secondary emails being different,
  - if specified emails obey semantical correctness,
- organizational unit config:
  - duplicated org units,
  - correct parent org unit path,
  - correct org unit path,
- groups config:
  - duplicated groups (based on group email),
  - if specified group email obeys semantical correctness,
  - valid group members roles,
  - valid group members emails.

In order to validate the file, run *GMan* with the `-validate` flag. In this case, the private
key and impersonated email can be omitted.

```bash
$ gman \
    -users-config myconfig.yaml \
    -groups-config myconfig.yaml \
    -orgunits-config myconfig.yaml \
    -validate
2020/06/17 19:24:49 ✓ Configuration is valid.
```

If the config is valid, the program exits with code 0, otherwise with a non-zero code. If this flag
is specified, *GMan* performs **only** the config validation. Otherwise, validation takes place
before every synchronization.

### Synchronizing

Synchronizing means updating GSuite's state to match the given configuration file. Without
specifying the `-confirm` flag the changes are just previewed:

```bash
$ gman \
    -private-key MYKEY.json \
    -impersonated-email me@example.com \
    -users-config myconfig.yaml \
    -groups-config myconfig.yaml \
    -orgunits-config myconfig.yaml
2020/06/25 18:55:54 ✓ Configuration is valid.
2020/06/25 18:55:54 ► Updating organization myorganization...
2020/06/25 18:55:54 ⇄ Syncing organizational units
2020/06/25 18:55:56 ✁ There is no org units to delete.
2020/06/25 18:55:56 ✎ There is no org units to create.
2020/06/25 18:55:56 ✎ There is no org units to update.
2020/06/25 18:55:56 ⇄ Syncing users
2020/06/25 18:55:56 ✁ There is no users to delete.
2020/06/25 18:55:56 ✎ There is no users to create.
2020/06/25 18:55:56 ✎ There is no users to update.
2020/06/25 18:55:56 ⇄ Syncing groups
2020/06/25 18:55:57 ✁ There is no groups to delete.
2020/06/25 18:55:57 ✎ There is no groups to create.
2020/06/25 18:55:57 ✎ There is no groups to update.
2020/06/25 18:55:57 ⚠ Run again with -confirm to apply the changes above.
```

Run the same command again with `-confirm` to perform the changes.

### Static Passwords

GMan can be used to manage dummy/testing accounts with predefined passwords. Note that you should never
put real passwords in cleartext anywhere near GMan, but if you have public passwords, e.g. for workshops
and demonstrations, this feature can be handy.

To make use of this, set a cleartext password for a user in your `users.yaml`:

```yaml
organization: myorganization
users:
  - givenName: Josef
    familyName: K
    primaryEmail: josef@myorganization.com
    orgUnitPath: /Developers
    password: i-am-not-secure-at-all
```

You must then opt-in to this feature by running GMan with `-insecure-passwords`:

```bash
$ gman \
    -private-key MYKEY.json \
    -impersonated-email me@example.com \
    -users-config myconfig.yaml \
    -groups-config myconfig.yaml \
    -orgunits-config myconfig.yaml \
    -insecure-passwords
2020/06/25 18:55:54 ✓ Configuration is valid.
2020/06/25 18:55:54 ► Updating organization myorganization...
...
```

GMan will now set the configured password and store its SHA256 hash as a custom schema field on the user.
On the next run, GMan will compare the hash with the configured password and update the user in GSuite
only if needed.

## Limitations

### Sending the login info email to the new users

Due to the fact that it is impossible to automate the send out of the login information
email via Google API there are two possibilities to enable the first log in of the new users:

- manually send the login information email from admin console via _RESET PASSWORD_ option
  (follow instructions on [this official Google documentation](https://support.google.com/a/answer/33319?hl=en))
- set up a password recovery for users (follow [this official Google documentation](https://support.google.com/a/answer/33382?p=accnt_recovery_users&visit_id=637279854011127407-389630162&rd=1&hl=en) to perform it).
  This requires the `recoveryEmail` field to be set for the users. Hence, in the onboarding
  message the new users ought to be informed about their new GSuite email address and that on
  the first login, the _Forgot password?_ option should be chosen, so the verification code
  can be sent to to the private recovery email.

### API requests quota

In order to retrieve information about licenses of each user, there are multiple API requests
performed. This can result in hitting the maximum limit of allowed calls per 100 seconds. In
order to avoid it, *GMan* waits after every Enterprise Licensing API request for 0.5 second.
This delay can be changed by starting the application with specified flag
`-throttle-requests <value>`, where value designates the waiting time in seconds (e.g. `5s`).

## Changelog

For detailed information on the latest updates, check the [Changelog](CHANGELOG.md).

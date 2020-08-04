# gman

*gman links all GSuite accounts with the matching user-list storage in form of a YAML. It is based on the [Aquayman](https://github.com/kubermatic-labs/aquayman) tool.*

**Features:**

- declare your users, groups and org units in code (IaC) via config YAML, which will then be applied to GSuite organization
- export the current state as a starter config file
- preview any action taken (validation)

**Table of contents:**
<!-- TOC -->
- [gman](#gman)
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
  - [Limitations](#limitations)
    - [Sending the login info email to the new users](#sending-the-login-info-email-to-the-new-users)
  - [Changelog](#changelog)
<!-- /TOC -->

## Installation

The official releases can be found [here](https://github.com/kubermatic-labs/gman/releases).

## Configuration & Authentication 

### Basics: Admin & Directory API 

The Directory API is intended for management of devices, groups, group members, organizational units and users. 

To be able to use it, please make sure that you have access to an admin account in the Admin Console and you have set up your API. 
For more detailed information, see [the official Google documentation](https://developers.google.com/admin-sdk/directory/v1/guides/prerequisites).

Moreover, to access the extended settings of the groups, the Groups Settings API must be enabled as well (see [the official documentation](https://developers.google.com/admin-sdk/groups-settings/prerequisites#prereqs-enableapis)).


### Service account 

To authorize and to perform the operations on behalf of *Gman* a Service Account is required. 
After creating one, it needs to be registered as an API client and have enabled this OAuth scopes: 

* https://www.googleapis.com/auth/admin.directory.user
* https://www.googleapis.com/auth/admin.directory.orgunit
* https://www.googleapis.com/auth/admin.directory.group
* https://www.googleapis.com/auth/admin.directory.group.member
* https://www.googleapis.com/auth/apps.groups.settings
* https://www.googleapis.com/auth/admin.directory.resource.calendar
  
Those scopes can be added in Admin console under *Security -> API Controls -> Domain-wide Delegation*.

Furthermore, please, generate a Key (save the *.json* config) for this Service Account. 
For more detailed information, follow [the official instructions](https://developers.google.com/admin-sdk/directory/v1/guides/delegation#create_the_service_account_and_credentials).

The Service Account private key must be provided to *Gman*. There are two ways to do so: 

- set up environmental variable: `GMAN_SERVICE_ACCOUNT_KEY=<VALUE>` 
- start the application with specified flag `-private-key </path/to/your/privatekey.json>`

### Impersonated Email 

Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore the service account needs to impersonate one of the admin users.

In order to delegate domain-wide authority to your service account follow this [official guide](https://developers.google.com/admin-sdk/directory/v1/guides/delegation#delegate_domain-wide_authority_to_your_service_account).

The impersonated email must be specified in *Gman*. There are two ways to do so: 

- set up environmental variable: `GMAN_IMPERSONATED_EMAIL=<VALUE>` 
- start the application with specified flag `-impersonated-email <value>`

### Config YAML 

All configuration of the users happens in a YAML file. See the annotated [config.example.yaml](/config.example.yaml) for more information.
This file must be created beforehand with the minimal configuration, i.e. organization name specified. 
In order to get the initial config of the users that are already in place in your Organizaiton, run *Gman* with `-export` flag specified, so the depicted on your side YAML can be populated. 

There are two ways to specify the path to the configuration YAML file:

- set up environmental variable: `GMAN_CONFIG_FILE=<VALUE>` 
- start the application with specified flag `-config <value>`

## Usage

After the completion of the steps above, *Gman* can perform for you: 

1. [exporting](#exporting) existing users in the domain;
2. [validating](#validating) the config file;
3. [synchronizing](#synchronizing) (comparing the state) the users without executing the changes;
4. and [confirming synchronization](#confirming-synchronization).

### Exporting

To get started, *Gman* can export your existing GSuite users into a configuration file.
For this to work, prepare a fresh configuration file and put your organisation name in it.
You can skip everything else:

```yaml
organization: myorganization
```

Now run *Gman* with the `-export` flag:

```bash
$ gman -config myconfig.yaml -export
2020/06/25 18:54:56 ► Exporting organization myorganization...
2020/06/25 18:54:56 ⇄ Exporting OrgUnits from GSuite...
2020/06/25 18:54:57 ⇄ Exporting users from GSuite...
2020/06/25 18:54:57 ⇄ Exporting groups from GSuite...
2020/06/25 18:54:58 ✓ Export successful.
```

Afterwards, the `myconfig.yaml` will contain an exact representation of your organizational unit, users and groups:

```yaml
organization: myorganization
org_units:
  - name: Developers
    description: dedicated org unit for devs
    parentOrgUnitPath: / 
    org_unit_path: /Developers 
users:
  - given_name: Josef
    family_name: K
    primary_email: josef@myorganization.com
    secondary_email: josef@privatedomain.com
    org_unit_path: /Developers
  - given_name: Gregor
    family_name: Samsa
    primary_email: gregor@myorganization.com
    secondary_email: gregor@privatedomain.com
    org_unit_path: /
groups:
  - name: Team Gman
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
  
In order to validate the file, run *Gman* with the `-validate` flag:

```bash
$ gman -config myconfig.yaml -validate
2020/06/17 19:24:49 ✓ Configuration is valid.
```

If the config is valid, the program exits with code 0, otherwise with a non-zero code.
If this flag is specified, *Gman* performs **only** the config validation. Otherwise, validation takes place before every synchronization. 

### Synchronizing

Synchronizing means updating GSuite's state to match the given configuration file. Without specifying the `-confirm` flag the changes are not performed:

```bash
$ gman -config myconfig.yaml
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

### Confirming synchronization

When running *Gman* with the `-confirm` flag the magic of synchronization happens!

 - The users, groups and org units - that have been depicted to be present in config file, but not in GSuite - are automatically created:

```bash
$ gman -config myconfig.yaml -confirm
2020/06/25 18:59:47 ✓ Configuration is valid.
2020/06/25 18:59:47 ► Updating organization myorganization...
2020/06/25 18:59:47 ⇄ Syncing organizational units
2020/06/25 18:59:48 ✎ Creating...
2020/06/25 18:59:49     + org unit: NewOrgUnit
2020/06/25 18:59:49 ⇄ Syncing users
2020/06/25 18:59:50 ✎ Creating...
2020/06/25 18:59:51     + user: someonenew@myorganization.com
2020/06/25 18:59:51 ⇄ Syncing groups
2020/06/25 18:59:51 ✎ Creating...
2020/06/25 18:59:54     + group: NewGroup
2020/06/25 18:59:54 ✓ Organization successfully synchronized.
```

- The users, groups and org units - that hold different values in the config file, than they have in GSuite - are automatically updated:

```bash
gman -config myconfig.yaml -confirm
2020/06/25 19:01:33 ✓ Configuration is valid.
2020/06/25 19:01:33 ► Updating organization myorganization...
2020/06/25 19:01:33 ⇄ Syncing organizational units
2020/06/25 19:01:34 ✎ Updating...
2020/06/25 19:01:35     ~ org unit: NewOrgUnit 
2020/06/25 19:01:35 ⇄ Syncing users
2020/06/25 19:01:36 ✎ Updating...
2020/06/25 19:01:36     ~ user: someonenew@myorganization.com
2020/06/25 19:01:36 ⇄ Syncing groups
2020/06/25 19:01:37 ✎ Updating...
2020/06/25 19:01:38     ~ group: UpdatedGroup
2020/06/25 19:01:38 ✓ Organization successfully synchronized.
```

- The users, groups and org units - that are present in GSuite, but not in config file - are automatically deleted:

```bash
$ gman -config myconfig.yaml -confirm
2020/06/25 19:06:04 ✓ Configuration is valid.
2020/06/25 19:06:04 ► Updating organization myorganization...
2020/06/25 19:06:04 ⇄ Syncing organizational units
2020/06/25 19:06:06 ✁ Deleting...
2020/06/25 19:06:07     - org unit: NewOrgUnit
2020/06/25 19:06:07 ⇄ Syncing users
2020/06/25 19:06:08 ✁ Deleting...
2020/06/25 19:06:08     - user: someonenew@myorganization.com
2020/06/25 19:06:08 ⇄ Syncing groups
2020/06/25 19:06:09 ✁ Deleting...
2020/06/25 19:06:10     - group: test group
2020/06/25 19:06:11     - group: UpdatedGroup
2020/06/25 19:06:11 ✓ Organization successfully synchronized.
```

## Limitations

### Sending the login info email to the new users

Due to the fact that it is impossible to automate the send out of the login information email via Google API there are two possibilities to enable the first log in of the new users: 

- manually send the login information email from admin console via _RESET PASSWORD_ option (follow instructions on [this official Google documentation](https://support.google.com/a/answer/33319?hl=en))
- set up a password recovery for users (follow [this official Google documentation](https://support.google.com/a/answer/33382?p=accnt_recovery_users&visit_id=637279854011127407-389630162&rd=1&hl=en) to perform it). *Gman* sets te secondary email address as a recovery email; hence, in your onboarding message you should inform the users about their new GSuite email address and that on the first login, the _Forgot password?_ option should be chosen, so the verification code can be sent to to the private secondary email. 

## Changelog

For detailed information on the latest updates, check the [Changelog](CHANGELOG.md).


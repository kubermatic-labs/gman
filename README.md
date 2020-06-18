# gman

*gman links all Gsuite accounts with the matching user-list storage in form of a YAML. It is based on the `Aquayman` tool.*

**Features:**

- declare your users in code (IaC) via YAML, which will then be applied to GSuite organization
- exports the current state as a starter config file
- preview of any action taken (validation)

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

.....

## Configuration & Authentication 

### Basics: Admin & Directory API 

The Directory API is intended for management of devices, groups, group members, organizational units and users. 

To be able to use it, please make sure that you have access to an admin account in the Admin Console and you have set up your API. 
For more detailed information, see [the official Google documentation](https://developers.google.com/admin-sdk/directory/v1/guides/prerequisites).

### Service account 

To authorize and to perform the operations on behalf of *Gman* a Service Account is required. 

Please, create one and generate a Key (save the *.json* config). 
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

To get started, *Gman* can export your existing Gsuite users into a configuration file.
For this to work, prepare a fresh configuration file and put your organisation name in it.
You can skip everything else:

```yaml
organization: myorganization
```

Now run *Gman* with the `-export` flag:

```bash
$ gman -config myconfig.yaml -export
2020/06/17 19:17:28 ► Exporting organization myorganization...
2020/06/17 19:17:28 ⇄ Exporting users from GSuite...
2020/06/17 19:17:28 ✓ Export successful.
```

Afterwards, the `myconfig.yaml` will contain an exact representation of your users:

```yaml
organization: myorganization
users:
    - given_name: Josef
      family_name: K
      primary_email: josef@myorganization.com
      secondary_email: josef@privatedomain.com
    - given_name: Gregor
      family_name: Samsa
      primary_email: gregor@myorganization.com
      secondary_email: gregor@privatedomain.com
```

### Validating

It's possible to validate a configuration file for:

- duplicated users (based on primary email),
- primary and secondary emails being different,
- if specified emails obey semantical correctness.
  
In order to validate the file, run *Gman* with the `-validate` flag:

```bash
$ gman -config myconfig.yaml -validate
2020/06/17 19:24:49 ✓ Configuration is valid.
```

If the config is valid, the program exits with code 0, otherwise with a non-zero code.
If this flag is specified, *Gman* performs **only** the config validation. Otherwise, validation takes place before every synchronization. 

### Synchronizing

Synchronizing means updating Gsuite's users state to match the given configuration file. Without specifying the `-confirm` flag the changes are not performed:

```bash
$ gman -config myconfig.yaml
2020/06/17 19:32:28 ✓ Configuration is valid.
2020/06/17 19:32:28 ► Updating organization Loodse…
2020/06/17 19:32:28 ⇄ Syncing users
2020/06/17 19:32:29 ✁ There is no users to delete.
2020/06/17 19:32:29 ✎ Found users to create: 
2020/06/17 19:32:29     + Someone New
2020/06/17 19:32:29 ✎ There is no users to update.
2020/06/17 19:32:29 ⚠ Run again with -confirm to apply the changes above.
```

### Confirming synchronization

When running *Gman* with the `-confirm` flag the magic of synchronization happens!

 - The users - that have been depicted to be present in config file, but not in Gsuite - are automatically created:

```bash
$ gman -config myconfig.yaml -confirm
2020/06/17 19:34:13 ✓ Configuration is valid.
2020/06/17 19:34:13 ► Updating organization myorganization…
2020/06/17 19:34:13 ⇄ Syncing users
2020/06/17 19:34:14 ✁ There is no users to delete.
2020/06/17 19:34:14 ✎ Found users to create: 
2020/06/17 19:34:14     + Someone New
2020/06/17 19:34:14 ✎ There is no users to update.
2020/06/17 19:34:14 Created user: someone@myorganization.training 
2020/06/17 19:34:14 ✓ Users successfully synchronized.
```

- The users - that are present in Gsuite, but not in config file - are automatically deleted:

```bash
gman -config myconfig.yaml -confirm
2020/06/17 19:37:39 ✓ Configuration is valid.
2020/06/17 19:37:39 ► Updating organization Loodse…
2020/06/17 19:37:39 ⇄ Syncing users
2020/06/17 19:37:39 ✁ Found users to delete: 
2020/06/17 19:37:39     - Someone ToDelete
2020/06/17 19:37:39 ✎ There is no users to create.
2020/06/17 19:37:39 ✎ There is no users to update.
2020/06/17 19:37:40 Deleted user: someone.new@loodse.training 
2020/06/17 19:37:40 ✓ Users successfully synchronized.
```

- The users - that hold different values in the config file, than they have in Gsuite - are automatically updated:

```bash
$ gman -config myconfig.yaml -confirm
2020/06/17 19:36:12 ✓ Configuration is valid.
2020/06/17 19:36:12 ► Updating organization Loodse…
2020/06/17 19:36:12 ⇄ Syncing users
2020/06/17 19:36:13 ✁ There is no users to delete.
2020/06/17 19:36:13 ✎ There is no users to create.
2020/06/17 19:36:13 ✎ Found users to update: 
2020/06/17 19:36:13     ~ Someone ChangedName
2020/06/17 19:36:13 Updated user: someone.new@loodse.training 
2020/06/17 19:36:13 ✓ Users successfully synchronized.
```

## Limitations

### Sending the login info email to the new users

Due to the fact that it is impossible to automate the send out of the login information email via Google API there are two possibilities to enable the first log in of the new users: 

- manually send the login information email from admin console via _RESET PASSWORD_ option (follow instructions on [this official Google documentation](https://support.google.com/a/answer/33319?hl=en))
- set up a password recovery for users (follow [this official Google documentation](https://support.google.com/a/answer/33382?p=accnt_recovery_users&visit_id=637279854011127407-389630162&rd=1&hl=en) to perform it). *Gman* sets te secondary email address as a recovery email; hence, in your onboarding message you should inform the users about their new Gsuite email address and that on the first login, the _Forgot password?_ option should be chosen, so the verification code can be sent to to the private secondary email. 

## Changelog

For detailed information on the latest updates, check the [Changelog](CHANGELOG.md).


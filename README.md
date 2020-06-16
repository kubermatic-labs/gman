# gman

*gman links all Gsuite accounts with the matching user-list storage in form of a YAML. It is based on the `Aquayman` tool.*

**Features:**

- declare your users in code (IaC) via YAML, which will then be applied to GSuite organization
- exports the current state as a starter config file
- preview of any action taken (validation)

<!-- TOC -->
- [gman](#gman)
  - [Installation](#installation)
  - [Configuration & Authentication](#configuration--authentication)
    - [Basics: Admin & Directory API](#basics-admin--directory-api)
    - [Service account](#service-account)
    - [Impersonated Email](#impersonated-email)
    - [Config YAML](#config-yaml)
  - [Usage](#usage)
  - [Changelog](#changelog)
<!-- /TOC -->



## Installation

.....

## Configuration & Authentication 
### Basics: Admin & Directory API 
The Directory API is intended for management of devices, groups, group members, organizational units and users. 

To be able to use it, please make sure that you have access to an admin account in the Admin Console and you have set up your API. 
For more detailed information, see [the official Google docs](https://developers.google.com/admin-sdk/directory/v1/guides/prerequisites).

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

...

## Changelog

For detailed information on the latest updates, check the [Changelog](CHANGELOG.md).

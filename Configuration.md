# Configuration

*Here can be found all the configuration details of GMan.*

**Table of contents:**
<!-- TOC -->
- [Configuration](#configuration)
  - [Organizational Units](#organizational-units)
  - [Users](#users)
    - [User Licenses](#user-licenses)
  - [Groups](#groups)
    - [Group's Permissions](#groups-permissions)
      - [Contacting owner](#contacting-owner)
      - [Viewing membership](#viewing-membership)
      - [Approving membership](#approving-membership)
      - [Posting messages](#posting-messages)
      - [Joining group](#joining-group)
<!-- /TOC -->

## Organizational Units

The organizational units are specified as the entries of the `org_units` collection.

Each OU contains:

| parameter    | type   | description  | required |
|--------------|--------|--------------|----------|
| name    | string | The name of the organizational unit. Inside of the OU's path it is the last entry, i.e. an organizational unit's name within the `/students/math/extended_math` parent path is `extended_math`  | yes |
| description  | string | The description of the organizational unit. | no  |
| parentOrgUnitPath | string | The organizational unit's parent path. If the OU is directly under the parent organization, the entry should contain a single slash  `/`. If OU is nested, then, for example, `/students/mathematics` is the parent path for  `extended_math` organizational unit with full path  `/students/math/extended_math`. | yes |
| org_unit_path  | string | The full path of the OU. It is derived from parentOrgUnitPath and organizational unit's name.   | yes |

## Users

The users are specified as the entries of the `users` collection.

Each user contains:

| parameter  | type   | description  | required |
|------------|--------|--------------|----------|
| given_name |  string  | first name of the user    | yes |
| family_name  |  string  | last name of the user  | yes |
| primary_email |  string  | a GSuite email address; must end with your domain name | yes   |
| secondary_email |  string    | additional, private email address   | no  |
| org_unit_path |  string    | org unit path indicates in which OU the user should be created; single slash '/' points to parent organization | no  |
| aliases  | list of strings | list of the user's alias email addresses   | no  |
| phones   | list of strings | list of the user's phone numbers    | no  |
| recovery_phone  |  string    | recovery phone of the user  | no  |
| recovery_email  |  string    | recovery email of the user; allows password recovery for the users  | no  |
| licenses | list of strings | Google products and related Stock Keeping Units (SKUs) assigned to the user; for detailed information about possible values, see table below | no  |
| employee_info: employee_ID   |  string    | employee ID  | no  |
| employee_info: department    |  string    | department | no  |
| employee_info: job_title  |  string    | title of the work position  | no  |
| employee_info: type  |  string    | description of the employment type  | no  |
| employee_info: cost_center   |  string    | cost center of the user's organization   | no  |
| employee_info: manager_email |  string    | email of the person (manager) the user is related to  | no  |
| location: building |  string    | building name   | no  |
| location: floor    |  string    | floor name/number    | no  |
| location: floor_section |  string    | floor section   | no  |
| addresses  |  string    | private address of the user    | no  |

### User Licenses

The user's licenses are the Google products and related Stock Keeping Units (SKUs).
The official list of all the available products can be found in [the official Google documentation](https://developers.google.com/admin-sdk/licensing/v1/how-tos/products).

GMan supports the following names as the equivalents of the Google SKUs:

| Google SKU Name (License) | GMan value |
|---------------------------|------------|
| G Suite Enterprise | GSuiteEnterprise |
| G Suite Business | GSuiteBusiness |
| G Suite Basic | GSuiteBasic
| G Suite Essentials | GSuiteEssentials |
| G Suite Lite | GSuiteLite |
| Google Apps Message Security | GoogleAppsMessageSecurity |
| G Suite Enterprise for Education | GSuiteEducation |
| G Suite Enterprise for Education (Student) | GSuiteEducationStudent |
| Google Drive storage 20 GB | GoogleDrive20GB |
| Google Drive storage 50 GB | GoogleDrive50GB |
| Google Drive storage 200 GB | GoogleDrive200GB |
| Google Drive storage 400 GB | GoogleDrive400GB |
| Google Drive storage 1 TB | GoogleDrive1TB |
| Google Drive storage 2 TB | GoogleDrive2TB |
| Google Drive storage 4 TB | GoogleDrive4TB |
| Google Drive storage 8 TB | GoogleDrive8TB |
| Google Drive storage 16 TB | GoogleDrive16TB |
| Google Vault | GoogleVault |
| Google Vault - Former Employee | GoogleVaultFormerEmployee |
| Cloud Identity Premium | CloudIdentityPremium |
| Google Voice Starter | GoogleVoiceStarter |
| Google Voice Standard | GoogleVoiceStandard |
| Google Voice Premier | GoogleVoicePremier |

Remark: *Cloud Identity Free Edition* is a site-wide SKU (applied at customer level), hence it cannot be managed by GMan as it is not assigned to individual users.

## Groups

The groups are specified as the entries of the `groups` collection.

Each user contains:

| parameter | type | description | required |
|-----------|------|-------------|----------|
| name  | string | name of the group |  yes   |
| email | string | email of the group; must end with your organization's domain name |    yes |
| description | string | group's description; max 300 characters |
|  who_can_contact_owner  | string | permissions to view contact owner of the group; for possible values see below |  yes |
|  who_can_view_members  | string | permissions to view group messages; for possible values see below |  yes |
|  who_can_approve_members  | string | permissions to approve members who ask to join groups; for possible values see below |  yes |
|  who_can_post  | string | permissions to post messages; for possible values see below |  yes |
|  who_can_join  | string | permissions to join group; for possible values see below |  yes |
| allow_external_members | bool | identifies whether members external to your organization can join the group | yes |
| is_archived | bool | allows the group content to be archived | yes |
| members | list of members | each member is specified by the email and the role; for the limits of numebr of users please refer to [the official Google documentation](https://support.google.com/a/answer/6099642?hl=en) | yes |
| member: email | string | primary email of the user | yes |
| member: role | string | role in the group of the user; possible values are: `MEMBER`, `OWNER` or `MANAGER` | yes |

### Group's Permissions

The group permissions designate who can perform which actions in the group.

#### Contacting owner

Permission to contact owner of the group via web UI. Field name is `who_can_contact_owner`. The entered values are case sensitive.

| possible value | description |
|----------------|-------------|
| ALL_IN_DOMAIN_CAN_CONTACT | all users in the domain |
| ALL_MANAGERS_CAN_CONTACT | only managers of the group |
| ALL_MEMBERS_CAN_CONTACT | only members of the group |
| ANYONE_CAN_CONTACT | any Internet user  |

#### Viewing membership

Permissions to view group members. Field name is `who_can_view_members`. The entered values are case sensitive.

| possible value | description |
|----------------|-------------|
| ALL_IN_DOMAIN_CAN_VIEW | all users in the domain |
| ALL_MANAGERS_CAN_VIEW | only managers of the group |
| ALL_MEMBERS_CAN_VIEW | only members of the group |
| ANYONE_CAN_VIEW | anyone in the group |

#### Approving membership

Permissions to approve members who ask to join group. Field name is `who_can_approve_members`. The entered values are case sensitive.

| possible value | description |
|----------------|-------------|
| ALL_OWNERS_CAN_APPROVE | only owners of the group |
| ALL_MANAGERS_CAN_APPROVE | only managers of the group |
| ALL_MEMBERS_CAN_APPROVE | only members of the group |
| NONE_CAN_APPROVE | noone in the group |

#### Posting messages

Permissions to post messages in the group. Field name is `who_can_post`. The entered values are case sensitive.

| possible value | description |
|----------------|-------------|
| NONE_CAN_POST | the group is disabled and archived; 'is_archived' must be set to true, otherwise will result in an error |
| ALL_MANAGERS_CAN_POST | only managers and owners of the group |
| ALL_MEMBERS_CAN_POST | only members of the group |
| ALL_OWNERS_CAN_POST | only owners of the group |
| ALL_IN_DOMAIN_CAN_POST | anyone in the organization |
| ANYONE_CAN_POST | any Internet user who can access your Google Groups service |

#### Joining group

Permissions to join the group. Field name is `who_can_join`. The entered values are case sensitive.

| possible value | description |
|----------------|-------------|
| ANYONE_CAN_JOIN | any Internet user who can access your Google Groups service |
| ALL_IN_DOMAIN_CAN_JOIN |  anyone in the organization |
| INVITED_CAN_JOIN | only invited candidates |
| CAN_REQUEST_TO_JOIN | non-members can request an invitation to join |

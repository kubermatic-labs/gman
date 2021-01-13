# Configuration

*Here are the configuration details of GMan.*

**Table of contents:**
<!-- TOC -->
- [Configuration](#configuration)
  - [Organizational Units](#organizational-units)
  - [Users](#users)
    - [User Licenses](#user-licenses)
  - [Groups](#groups)
<!-- /TOC -->

## Organizational Units

The organizational units (OU) are specified as the entries of the `orgUnits` collection.

```yaml
organization: exampleorg
orgUnits:
  - # unique name (required)
    name: Org Unit 1
    description: An optional description text.
    # The organizational unit's parent path.
    # If the OU is directly under the parent organization, the entry should contain a single slash `/`
    # (which is also the default)
    parentOrgUnitPath: /
    blockInheritance: false

  - ...
```

## Users

The users are specified as the entries of the `users` collection.

```yaml
organization: exampleorg
users:
  - # first name of the user (required)
    givenName: Roxy
    # last name of the user (required)
    familyName: Sampleperson
    # a GSuite email address; must end with your domain name (required)
    primaryEmail: roxy@example.com
    # org unit path indicates in which OU the user should be created;
    # single slash '/' points to parent organization (default)
    orgUnitPath: /AwesomePeople
    # optional list of additional email aliases
    aliases:
      - roxyrocks@example.com
    # optional list of phone numbers
    phones:
      - 555-887951-87
    # recovery phone number (optional)
    recoveryPhone: 555-887951-87
    # recovery email address (optional)
    recoveryEmail: roxys-recovery-address@gmail.com
    # list of licenses this user is assigned to
    licenses:
      - GoogleDriveStorage20GB
      - GoogleVoicePremier
    # optional detailed employee information
    employee:
      # employee ID
      id: ''
      department: ''
      jobTitle: ''
      type: ''
      costCenter: ''
      managerEmail: ''
    # optional location info
    location:
      building: ''
      floor: ''
      floorSection: ''
    # optional address
    address: "Rue d'Example 42, 12345 Sampleville"

  - ...
```

### User Licenses

The user's licenses are the Google products and related Stock Keeping Units (SKUs).
The official list of all the available products can be found in [the official Google documentation](https://developers.google.com/admin-sdk/licensing/v1/how-tos/products).

GMan has a list of licenses built-in, but this can be overwritten by running gman with
`-licenses-config <file>`, which must be a YAML file that contains a list of licenses.
Run GMan with `-licenses` to see the list of default licenses. If you also specify
`-licenses-yaml`, you get an output that can be directly used as a config file.

Remark: *Cloud Identity Free Edition* is a site-wide SKU (applied at customer level),
hence it cannot be managed by GMan as it is not assigned to individual users.

## Groups

The groups are specified as the entries of the `groups` collection.

```yaml
organization: exampleorg
groups:
  - # unique name (required)
    name: Christmas 2021
    # group email address (required)
    email: christmas2021@example.com

    # the following settings control access to the group;
    # the shown value is the implicit default value

    # one of ALL_MANAGERS_CAN_CONTACT, ALL_MEMBERS_CAN_CONTACT, ALL_IN_DOMAIN_CAN_CONTACT, ANYONE_CAN_CONTACT
    whoCanContactOwner: ALL_MANAGERS_CAN_CONTACT
    # one of ALL_MANAGERS_CAN_VIEW, ALL_MEMBERS_CAN_VIEW, ALL_IN_DOMAIN_CAN_VIEW
    whoCanViewMembership: ALL_MEMBERS_CAN_VIEW
    # one of ALL_MANAGERS_CAN_APPROVE, ALL_OWNERS_CAN_APPROVE, ALL_MEMBERS_CAN_APPROVE, NONE_CAN_APPROVE
    whoCanApproveMembers: ALL_MANAGERS_CAN_APPROVE
    # one of NONE_CAN_POST, ALL_OWNERS_CAN_POST, ALL_MANAGERS_CAN_POST, ALL_MEMBERS_CAN_POST, ALL_IN_DOMAIN_CAN_POST, ANYONE_CAN_POST
    whoCanPostMessage: ALL_MEMBERS_CAN_POST
    # one of INVITED_CAN_JOIN, CAN_REQUEST_TO_JOIN, ALL_IN_DOMAIN_CAN_JOIN, ANYONE_CAN_JOIN
    whoCanJoin: INVITED_CAN_JOIN

    # whether external users can join the group
    allowExternalMembers: false

    # whether the group is archived (readonly)
    isArchived: false

    # list of members in this group
    members:
      - email: roxy@example.com
      - email: rubert@example.com
      - email: santa@northpole.example.com
        # each member must be either OWNER, MANAGER or MEMBER (default)
        role: OWNER

  - ...
```

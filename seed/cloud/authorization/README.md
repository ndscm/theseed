# Authorization: Identity, Group, Role

The ndscm authorization system controls what each caller is allowed to do. It
builds on four concepts — permissions, identities, groups, and roles — to
provide fine-grained access control across every service in the platform.

Every authenticated request must declare who is calling (identity) and which
role they are acting under. The system resolves the caller's group memberships,
validates the assumed role, and passes the resulting permission set to the
service.

## Terms

### Permission

A tuple of `(service:operation:resource)`. Each component may contain wildcards.
It is the downstream service's responsibility to interpret the operation and
resource segments. The tuple `(*:*:*)` represents super-admin access.

### Handle

A handle is the unique identifier for an entity within a given domain — the part
before the `@` in an email-style address (e.g. `christina` in
`christina@ndscm.com`).

Handles in ndscm must follow these rules:

- Only lowercase ASCII letters (`a-z`) and digits (`0-9`) are allowed.
- Words are separated by a single dash (`-`). No other punctuation or whitespace
  is permitted.

> The sub-mailbox suffix (`+topic`) is not part of the handle. For example,
> `christina+ci@ndscm.com` still resolves to the handle `christina`.

Handles appear in identities, groups, and roles. Naming conventions for each
type are described in the sections below.

### Identity

A credential recognized by the authentication system. An identity can represent:

- A general human user account
- A general agent user account
- A user account with specific scenario
- A general service account
- A service account with deployment-tier

Every identity is scoped by a handle with a domain, similar to an email address.
For example:

- `nagi@ndscm.com` — human user account
- `christina@ndscm.com` — agent user account
- `nagi-cli@ndscm.com` — user identity for CLI usage
- `admin-nagi@ndscm.com` — admin identity for the nagi use case
- `golink@ndscm.com` — service account
- `golink-prod@ndscm.com` — service account for the prod tier

The domain can be omitted when operating within a known domain, e.g. `nagi@`,
`golink-prod@`. An alternative form places the `@` symbol before the handle, but
this is rarely used within ndscm.

Every identity is also assigned a mailbox and an IM account for communication.

### Group

A collection of identities or other groups.

Every group is scoped by a group handle with a domain, similar to an email
address. For example:

- `nagi-team@ndscm.com` — nagi team members
- `nagi-direct@ndscm.com` — nagi direct reports
- `golink-team@ndscm.com` — golink service maintainers
- `golink-user@ndscm.com` — golink service users (typically used as a
  notification mailing list)
- `golink-support@ndscm.com` — golink service support channel

Every group must be attached to an identity, which means the group handle always
contains at least one dash.

Every group is also assigned a mailing list and an IM channel for communication.

### Role

A named collection of permissions. A role can be assigned to a group or to an
individual identity.

Every role is scoped by a role handle with a domain, similar to an email
address. The handle must end with `-role`. For example:

- `admin-role@ndscm.com`
- `golink-prod-role@ndscm.com`
- `golink-sre-role@ndscm.com`

An identity must perform operations on behalf of exactly one role. A common (but
not required) convention is to add a `-role` suffix to an identity or group
handle, binding that role specifically to that identity or group.

## Evaluation

A permission-scoped operation is evaluated as:

```
Identity ${identity} (in group ${group}) requests service ${service} on behalf of role ${role}
```

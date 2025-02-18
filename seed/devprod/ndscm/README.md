---
author: @nagi
link: go/ndscm
created: 2024-10-10
updated: 2025-02-27
---

# NDSCM

## Overview

NDSCM (Not Distributed Source Code Management) is a monorepo scm inspired by
Google Piper, which is known as the google source code management system.
Comparing to distributed SCM systems like git, it is designed for meeting the
needs of enterprise usage to avoid code leak and collect all code and
development results in one central place.

## Advantages

**Advantages for better human control**

- Employees and contractors only download the code file that they need to view
  and edit. If too many files are touched by a single employee, the system will
  block the request and raise an alert.
- Trust the system for permission control instead of distributing different
  repos by hand, which reduce the chance of misconfiguration of permissions.

**Advantages for monorepo**

- All the code in a company is stored in one place, which helps teams to reuse
  wheels made by co-workers without caring about the wheel versions
- Dependencies across different programming languages can be easily compiled and
  tested (with Bazel).
- Build cache can be easily shared between developers (with Bazel).
- External dependencies can get more frequently upgraded in a monorepo, which
  reduce the chances of releasing new versions with security risks.

**Advantages for accelerating development productivity**

- Stop code review process for being a blocker of developing features based on a
  pending review patch

## Backgroud

Git is the most popular source code management (SCM) system across the world
since 2004. It was originally designed for linux team with members in different
companies. However, it was never adopted by giants like google or facebook.
These companies needs a stronger system to control when and which employees can
see what than git. So google created piper SCM (internally called google3, means
the 3rd generation of google monorepo). Based on the concept of monorepo, they
also built and shared Blaze - the build system for monorepo.

Since the software engineering has became the main component of the economic
society and people's life. The developing and depolying process has changed a
lot since 21st centray. Cloud cluster management and code review are introduced
to keep the software product safer and business success. These beneficial new
tech stack slow down the traditional single-person-army software development
productivity and the traditional SCM design is not taking it into consideration.

According to the wikipedia page of scm, there isn't a more popular open source
scm since 2008. This is abnormal for devops field since so many new enviornment
and society changes are happening during the recent decade. So we are going to
design and build a new centralized source code management system for new
open-mind companies.

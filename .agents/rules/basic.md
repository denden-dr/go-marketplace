---
trigger: always_on
---

# File Naming Format
- dto `something_dto`
- test `something_test`
- handler `something_handler`
- middleware `something_middleware`
- service `something_service`
- repository `something_repo`

# Object Naming Format
- dto request/response `SomethingRequest` `SomethingResponse`
- service interface `SomethingService` object `somethingService`
- repository interface `SomethingRepo` object `somethingRepo`
- test `TestSomething_Fail`

# Common Guide
- dto object should not contain any domain model and sensitive attribute
- handler should only reference interface of services
- services should only reference interface of repositories
- unit test should isolated, discover every failed route first, placed on the same directory

# Planning Guide
Create planning with high level human instruction with no code generation which junior programmer / basic AI Model can understand.

# Documentation Guide
- If it's not exist, just skip Documentation,
- if it's exist, create planning to update documentation of current change 

# Finishing Plan Guide
- Make sure the app could run properly
- create _report.md at root project folder of what you did

# Command Guide
lookup to the Makefile first, if it's not defined or exist, do it manually
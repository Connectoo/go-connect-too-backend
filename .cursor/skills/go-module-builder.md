# Go Module Builder

Use when creating a new backend module.

Inputs:
- module name
- main entity
- required endpoints
- database tables needed

Rules:
- create files under internal/modules/[module_name]
- include handler.go, service.go, repository.go, model.go, dto.go, routes.go
- handlers only parse, validate, call service, return response
- services contain business logic
- repositories contain database logic only
- use context.Context
- use constructors
- run gofmt
- do not create unrelated modules

Output:
- list files created
- explain routes added
- explain next required migration or tests
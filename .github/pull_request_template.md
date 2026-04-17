## Summary

- what changed
- why it changed

## Validation

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `npm run docs:build` when docs changed

## Boundary check

- [ ] no app-specific REST DTOs were added to public packages
- [ ] no product-specific route contracts were added
- [ ] the change respects transport vs app-layer separation

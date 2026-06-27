## Description

Please include a summary of the change and the issue it addresses. List any dependencies that are required for this change.

Fixes # (issue number)

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] New Host Adapter (adds support for a new editor/agent tool)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update / cleanup

## How Has This Been Tested?

Please describe the tests that you ran to verify your changes. Provide instructions so we can reproduce.

- [ ] **Contract/Unit Tests**: `go test ./...` output and new cases added.
- [ ] **Fixture Validation**: If adding an adapter, verify `internal/adapters/fixtures/` contains test payloads.
- [ ] **Manual Verification**: Briefly explain manual testing (e.g. running daemon + hook client against live editor).

## Checklist

- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings or build errors
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] My commits are GPG-signed (run `git log --show-signature` to verify)

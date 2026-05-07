---
on:
  pull_request:
    types: [closed]
  workflow_dispatch: {}
if: github.event.pull_request.merged == true

timeout-minutes: 35

# Change ALL permissions to read. 
# Strict mode will fail if any 'write' is found here.
permissions:
  contents: read
  pull-requests: read

safe-outputs:
  create-pull-request:
    # This replaces the need for 'pull-requests: write'
    fallback-as-issue: false
    protected-files:
      policy: blocked
      exclude:
        - README.md
        - DOCS.md
---

# Update Documentation Agent

## Goal
Keep the project's documentation in README.md up-to-date with the codebase.

## Instructions
1. Check if the file named `README.md` contains the title `plentymarkets microservice go boilerplate`. If yes, it means it's a default file and it should be completely changed by our new documentation. If it doesn't contain the text, append documentation to the end.
2. The documentation must prioritize the "Data Flow" section, placing it near the top as it conveys the most relevant information. It should then cover the project's purpose, installation, configuration, and API.
3. If `README.md` exists, review the diff of the merged Pull Request and update `README.md` to incorporate the changes. Ensure the "Data Flow" section remains prominent and near the top.
4. Reduce the number of code examples. Configuration examples and other lengthy code blocks should be linked to their source files instead of being included directly in the documentation.
5. Always create a Pull Request with the new or updated `README.md`.

## PR Details
- **Title**: "feat(docs): add documentation to README.md based on PR #${{ github.event.pull_request.number }}"
-- **Branch**: "gh-aw/update-docs-${{ github.event.pull_request.number }}"
- **Body**: "Automated documentation sync following the merge of PR #${{ github.event.pull_request.number }}."

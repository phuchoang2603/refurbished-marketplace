# GitHub workflow

Issue templates: **Feature**, **Bug**, and **Chore**.

| Template    | Use for                                                       |
| ----------- | ------------------------------------------------------------- |
| **Feature** | New capability, infrastructure, or larger improvements        |
| **Bug**     | Broken behavior                                               |
| **Chore**   | Tooling, refactors, deps, maintenance without a product story |

Large work can be a single **Feature** issue with a detailed checklist, or split into several focused issues that reference each other. Use **OpenSpec** for design and task breakdown when the change is non-trivial — see [openspec.md](openspec.md).

## Labels

Apply **type** and **area** as GitHub labels when opening an issue (not in the issue body):

- `type: feature`, `type: bug`, `type: chore`
- `area: web`, `area: payment`, `area: orders`, `area: cart`, `area: products`, `area: users`, `area: infra`, `area: ci`, `area: observability`, `area: docs`

## Example flow

1. Open a **Feature**, **Bug**, or **Chore** issue using the template.
2. Optional: create a matching **OpenSpec change** for design/tasks.
3. Branch and implement; open a PR using the PR template (`Closes #<issue>`).
4. Merge when CI passes and the issue acceptance criteria are met.

## OpenSpec ↔ GitHub mapping

| OpenSpec         | GitHub                                            |
| ---------------- | ------------------------------------------------- |
| Change proposal  | Feature issue (optional but recommended)          |
| `tasks.md` items | Checklist in the issue or follow-up issues        |
| PR               | `Closes #<issue>`; OpenSpec change in PR template |

## Pull requests

- Use the PR template in `.github/pull_request_template.md`.
- Keep PRs focused; prefer one issue per PR when practical.
- Mention the OpenSpec change name when applicable.
- Describe how you verified the change (usually via Tilt).

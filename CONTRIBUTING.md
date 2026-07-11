<!-- TOC -->

- [Contributing](#contributing)
- [Tip](#tip)

<!-- TOC -->

# Contributing

Your contribution is very welcome!

Follow these steps whenever you want to improve this repository.

- Install the following packages: `git`, `go` (see `app/go.mod` for the minimum version), and a text editor of your choice. Run `make check-tools` to verify your machine has everything needed for development, build, test, and deploy.
- Fork this repository. See this tutorial: https://help.github.com/en/github/getting-started-with-github/fork-a-repo
- Configure your GitHub account to use SSH instead of HTTPS. Watch this tutorial to learn how to set it up: https://help.github.com/en/github/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account
- Clone the resulting fork to your computer.
- Add the upstream repository URL with the command below.

```bash
git remote -v
git remote add upstream git@github.com:aeciopires/mytoolkit.git
git remote -v
```

- Create a branch using the pattern:

```bash
git checkout -b BRANCH_NAME
```

- Make sure you're on the correct branch, using the command below.

```bash
git branch
```

- The branch in use is marked with a `*` before its name.
- Make the necessary changes.
- If you touched Go code under `app/`, run the checks before committing:

```bash
make fmt
make vet
make lint
make test
```

- If you touched the Helm chart, run `make helm-lint` and `make helm-docs` (regenerates `helm/mytoolkit/README.md` — never edit that file by hand; it also syncs `Chart.yaml`'s `appVersion` to the root `VERSION` file first, so don't hand-edit `appVersion` either).
- If you touched a tool's behavior, re-run every example in its `docs/api/<tool>.md` and `docs/cli/<tool>.md` against the real binary and update the docs to match — don't hand-type expected output or error messages. Also check its `## Workflow` Mermaid diagram still matches the real request lifecycle.
- If you touched a REST handler's request/response shape (or added a new tool), update its `swaggo/swag` annotations and run `make swagger-gen` (regenerates `app/docs/` — never edit those files by hand either). See `.skills/swagger/SKILL.md`.
- Commit your changes on the newly created branch, preferably one commit per edited/created file.
- Push the commits to the remote repository with the command:

```bash
git push --set-upstream origin BRANCH_NAME
```

- Open a Pull Request (PR) against the `main` branch of the original repository. See this [tutorial](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork).
- Update the content with the reviewer's suggestions (if needed).
- After your PR is approved and merged, update your local repository with the commands below.

```bash
git checkout main
git pull upstream main
```

- Remove the local branch after your PR is approved and merged, using the command:

```bash
git branch -d BRANCH_NAME
```

- Update the `main` branch of your local repository.

```bash
git push origin main
```

- Push the local branch deletion to your GitHub repository with the command:

```bash
git push --delete origin BRANCH_NAME
```

- To keep your fork in sync with the original repository, run these commands:

```bash
git pull upstream main
git push origin main
```

Reference:
- https://blog.scottlowe.org/2015/01/27/using-fork-branch-git-workflow/

## Versioning

The repo-root [`VERSION`](VERSION) file is the single source of truth for the application version — it's read by both `make build` (embedded into the Go binary, shown by `mytoolkit --version`/`-v`) and `make docker-build`/`docker-buildx`/`docker-push` (used as the image tag). If your change warrants a version bump, update `VERSION` and add an entry to `CHANGELOG.md` in the same PR.

## Publishing a Docker image

`make docker-push` builds and pushes a multi-arch (`linux/amd64` + `linux/arm64`) image to Docker Hub. It prompts interactively for your username, password/access token, and target repository — the password is read with hidden input and piped directly into `docker login --password-stdin`, so it's never echoed, logged, or written to disk. Prefer a Docker Hub [access token](https://docs.docker.com/security/for-developers/access-tokens/) over your account password. Only run this target yourself, from your own terminal — never script or automate it with embedded credentials.

# Tip

**You can use the text editor of your choice, whichever you feel most comfortable with.**

But VSCode (https://code.visualstudio.com), combined with the following plugins, helps the editing/review process, mainly by allowing content preview before commit, analyzing Markdown syntax, and generating the automatic summary as section titles are created/changed.

- Markdown-lint: https://marketplace.visualstudio.com/items?itemName=DavidAnson.vscode-markdownlint
- Markdown-toc: https://marketplace.visualstudio.com/items?itemName=AlanWalk.markdown-toc
- Markdown-all-in-one: https://marketplace.visualstudio.com/items?itemName=yzhang.markdown-all-in-one
- YAML: https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml
- Helm-intellisense: https://marketplace.visualstudio.com/items?itemName=Tim-Koehler.helm-intellisense
- Go: https://marketplace.visualstudio.com/items?itemName=golang.go
- GitLens: https://marketplace.visualstudio.com/items?itemName=eamodio.gitlens
- Themes for VSCode:
    - https://code.visualstudio.com/docs/getstarted/themes
    - https://dev.to/thegeoffstevens/50-vs-code-themes-for-2020-45cc
    - https://vscodethemes.com/

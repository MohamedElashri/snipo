# Releasing Snipo

Use one `vX.Y.Z` tag for the server, Snippy, VS Code extension, and both
browser extensions. Target **v2.0.0** next, then `v2.0.1` for patches.
Verify the chosen version exceeds every published store version.

## Before the first release

The workflow improvements below are still pending. Complete them before tagging:

- Make tests and Go/npm/container security checks block publication; add a
  preview that builds everything without publishing.
- Fix version/lockfile synchronization, Firefox source packaging, browser smoke
  tests, and Homebrew URLs to use `vX.Y.Z`.
- Configure store publishing credentials and jobs. Document supported Go
  installation methods for v2.

## Prepare a release PR

1. Fix security findings and update dependencies and vendored assets.
2. Synchronize product versions:

   ```bash
   ./scripts/bump-version.sh 2.0.0
   git diff
   ```

   Use no leading `v`. Verify the server and TUI version values, both browser
   manifests, both extension package files, and the root lockfile match.
   The current script does not update every entry.
3. Update the main and VS Code changelogs, including upgrade instructions.
4. Run all component tests and security checks, then merge the PR into `main`.

## Publish

Update `main`, run the preview on that commit, and install/test its packages:

```bash
git switch main
git pull --ff-only origin main
```

After preview passes, tag that exact commit without pulling additional changes:

```bash
git tag -a v2.0.0 -m "Snipo ecosystem 2.0.0"
git push origin v2.0.0
```

Once implemented, Actions will validate, build, and publish the GitHub release
and Docker images, submit the extensions, and update Homebrew.

Check each destination in the release summary; browser store reviews may still
be pending. Retry failed destinations using the original artifacts. For changed
code, prepare a new patch release; never move a published tag.

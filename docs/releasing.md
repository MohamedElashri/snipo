# Releasing Snipo

One `vX.Y.Z` tag releases the server, Snippy, VS Code extension, and both
browser packages. Use `v2.0.0` for the coordinated major release, then `v2.0.1`
for patches. All product versions must match; use no prerelease suffixes.

## Prepare

1. Fix security findings and update dependencies and vendored assets.
2. Run `./scripts/bump-version.sh 2.0.0`, review the diff, and update the main
   and VS Code changelogs with upgrade instructions.
3. Open and merge a release PR after all checks pass.

## Preview

In Actions, run **Unified Release** manually with the merged commit SHA as
`ref`. This runs tests, audits, image scans, and packaging without publishing.
Install and test the resulting `release-bundle` artifacts.

Required Actions secrets: `DHI_USERNAME`, `DHI_API_KEY`, and `TAP_GITHUB_TOKEN`.
GitHub releases and GHCR use the workflow's built-in token.

## Publish

Tag the exact `main` commit that passed preview:

```bash
git tag -a v2.0.0 <verified-commit-sha> -m "Snipo ecosystem 2.0.0"
git push origin v2.0.0
```

Actions publishes the GitHub release, standard/hardened Docker images, and
Homebrew formula. Check the run summary for each destination's result.

## Extension stores and retries

Store uploads are manual for now. Submit the release's VSIX to VS Code, the
Chrome ZIP to Chrome Web Store, and the Firefox ZIP plus source ZIP to Mozilla.
Firefox must sign the extension before it can be installed normally.

Use **Re-run failed jobs** to retry a failed destination using the retained
artifacts. Never move a published tag or replace published packages; changed
code needs a new patch release.

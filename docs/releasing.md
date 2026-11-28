# Releasing Snipo

One `vX.Y.Z` tag releases the server, Snippy, VS Code extension, and both
browser packages. All product versions must match; use no prerelease suffixes.
Preparing or merging a release PR does not
publish a release; pushing its version tag starts publication.

## Prepare

1. Fix security findings and update dependencies and vendored assets.
2. Run `./scripts/bump-version.sh 2.2.1`, review the diff, and update the main
   and relevant client changelogs with user-visible changes. Mark the version as
   **Unreleased** until publication is scheduled.
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
git tag -a v2.2.1 <verified-commit-sha> -m "Snipo ecosystem 2.2.1"
git push origin v2.2.1
```

Actions publishes the GitHub release, standard/hardened Docker images, and
Homebrew formula. Once the GitHub release is public, three independent jobs
publish the VSIX to Visual Studio Marketplace and submit the Chrome and Firefox
ZIPs to their stores. Check the run summary for each destination's result.

## Configure extension publishing

Create the `vscode-marketplace`, `chrome-web-store`, and `firefox-add-ons`
GitHub Actions environments before the first automated release. Allow `v*`
tags as deployment sources.

- For **VS Code**, in the `muhammadelashri` publisher's Trusted Publishing
  settings, register owner `MohamedElashri`, repository `snipo`, workflow
  `unified-release.yml`, and environment `vscode-marketplace`. The job uses
  GitHub OIDC. [Marketplace setup](https://code.visualstudio.com/api/working-with-extensions/publishing-extension#publish-with-github-actions-and-trusted-publishing)
- For **Chrome**, enable the Chrome Web Store API and link a Google Cloud service
  account to the existing Snipo publisher. Add `CHROME_SERVICE_ACCOUNT_JSON` as
  an environment secret and `CHROME_PUBLISHER_ID` as an environment variable or
  secret in `chrome-web-store`. The publishing script already has the item ID.
  [Service account setup](https://developer.chrome.com/docs/webstore/service-accounts)
- For **Firefox**, create AMO API credentials for an author of the existing Snipo
  listing and set `AMO_JWT_ISSUER` and `AMO_JWT_SECRET` as environment secrets in
  `firefox-add-ons`. The job includes the matching source ZIP for review.
  [AMO API credentials](https://addons.mozilla.org/en-US/developers/addon/api/key/)

Chrome and Firefox jobs report submission. Store review controls when updates
become available. Firefox signs the approved extension for normal installation.

## Retry a release

Use **Re-run failed jobs** to retry a failed destination using the retained
artifacts. If Chrome is already reviewing another version, retry after that
review finishes. Keep published tags and packages in place. Changed code needs
a new patch release.

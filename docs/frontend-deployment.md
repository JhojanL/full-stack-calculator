# Frontend deployment

[Deploy frontend](../.github/workflows/deploy-frontend.yml) publishes `frontend/dist` to GitHub Pages on pushes to `master` or a manual dispatch from `master`. The expected public URL is **https://calculator.jhojanlerma.dev**.

## Workflow

The build job uses Node `24.19.0` and pnpm `12.5.1`, installs with `pnpm --dir frontend install --frozen-lockfile`, and runs `pnpm --dir frontend validate` (checks, unit tests, and production build). Full-stack browser tests remain in pull-request CI. Official Pages Actions configure the site, upload the build artifact, and deploy it through the `github-pages` environment. Actions are pinned to commit SHAs. The build has only contents/Pages read permissions; deployment has only Pages write and OIDC token permissions. Deployment concurrency prevents overlapping runs without cancelling an active deployment.

## API configuration

The workflow sets the public build-time configuration `VITE_API_BASE_URL=https://calculator-api.jhojanlerma.dev`; no API URL secret or repository variable is needed. Production requests go to that origin's `/calculate` endpoint. Vite's base is `/` for the custom domain. Development always uses relative requests through the existing local proxy; production builds without an API base also use relative requests, preserving local preview and browser tests. Changing the deployed API URL requires rebuilding the frontend.

## GitHub Pages setup

In repository **Settings → Pages**, select **GitHub Actions** as the build and deployment source, retain the existing custom domain `calculator.jhojanlerma.dev`, and enable **Enforce HTTPS** when available. Ensure the `github-pages` environment permits deployments from `master`. If DNS is not already configured, point the `calculator` subdomain's DNS CNAME to `jhojanl.github.io`. Actions-based Pages deployments do not require a repository `CNAME` file; see [GitHub's custom-domain guidance](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site) and [Pages workflow requirements](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).

## Backend CORS

The backend must allow the exact frontend origin `https://calculator.jhojanlerma.dev` through its `-cors-trusted-origins` configuration and accept the JSON POST preflight. This workflow deploys only the frontend; see [backend deployment](backend-deployment.md) for server configuration.

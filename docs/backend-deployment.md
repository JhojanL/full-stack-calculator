# Backend deployment

The [deployment workflow](../.github/workflows/deploy-backend.yml) deploys pushes
to `master` and supports manual runs on `master`. It audits the backend using the
same commands/tool versions as CI, builds with `make -C backend build/api` and
`CGO_ENABLED=0`, and sends `backend/bin/linux_amd64/api` over verified SSH.
It does not wait for the separate PR-only CI workflow or run frontend checks.

The approved deployment target is `/srv/calculator/api`, run by
`calculator.service`, with `http://127.0.0.1:4001/healthcheck` for local verification.
These are setup requirements, not verified live server state. At implementation
time, `backend/remote/production/calculator.service` and `Caddyfile` were empty;
no installed calculator service, server access, or sudo policy was verified.

## One-time server prerequisites

An administrator must prepare a Linux x86_64 server with SSH on port 22, systemd
at `/usr/bin/systemctl`, sudo, curl, jq, and standard GNU utilities. Create a
dedicated deployment account and a separate unprivileged calculator runtime
account. The deployment account needs write/search access to `/srv/calculator`
and its parent directories must be real directories, not symlinks. The runtime
account needs read/execute access to the installed binary; it should not be able
to replace it. Keep the unit and any systemd overrides root-owned and outside
the deployment account's write access.

Install and enable `/etc/systemd/system/calculator.service` separately. Its
`ExecStart` must use `/srv/calculator/api -port=4001 -env=production`, with a
non-root `User` and suitable `Group` and `WorkingDirectory`. Configure bounded
start/stop timeouts (for example, 30 seconds) and any restart policy there.
Ensure port 4001 is available and assigned exclusively to calculator. The API
listens on all interfaces, so keep this port private with the server firewall.
The workflow does not create accounts, directories, units, or proxy routes.

Use `visudo` to grant only the exact restart command, replacing `DEPLOY_USER`
with the configured deployment username:

```sudoers
DEPLOY_USER ALL=(root) NOPASSWD: /usr/bin/systemctl restart calculator.service
```

Do not grant unrestricted `systemctl`, shell, or filesystem commands. As that
user, check `sudo -k -n -l /usr/bin/systemctl restart calculator.service`.
During a planned service restart, verify
`sudo -k -n /usr/bin/systemctl restart calculator.service` succeeds without a
password. Listing permission alone does not prove that execution is passwordless;
the workflow also uses noninteractive sudo for the actual restart and fails if
authentication is required. Cached credentials are ignored.

## GitHub configuration

Create the `backend-production` Environment and restrict deployments to `master`.
Configure:

| Setting | Kind | Value |
| --- | --- | --- |
| `DROPLET_SSH_KEY` | Secret | Deployment account's private SSH key |
| `DROPLET_HOST` | Variable | Server DNS name or IPv4 address |
| `DROPLET_USER` | Variable | Dedicated deployment account |
| `DROPLET_KNOWN_HOSTS` | Variable | Host's OpenSSH known-hosts entry verified through a trusted channel |

The corresponding public key must be authorized on the server. SSH must be
reachable from the runner. The workflow uses `StrictHostKeyChecking yes` and
batch authentication; it does not discover or accept unknown keys during a run.
See the [OpenSSH configuration reference](https://man.openbsd.org/ssh_config).

## Release and failure behavior

After server preflight, the workflow streams the binary to a unique temporary
file inside `/srv/calculator`, makes it executable, and atomically renames it
over `api`. Only `calculator.service` is restarted. Bounded retries require an
active service and health JSON reporting `available` and `production`. Checks
run on the server through SSH; they do not verify public DNS, TLS, or Caddy.

No Greenlight resources, databases, systemd units, or shared Caddy configuration
are uploaded or modified. Public calculator routing, if needed, is a separate
administrator task. No application secrets are required.

The workflow serializes calculator deployments without cancelling an active
run, has job/step/network timeouts, and removes its runner private key with
`if: always()`. GitHub behavior is described in the
[workflow syntax reference](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax).

A restart briefly interrupts service. Failed verification fails the workflow;
there is no automatic rollback or retained previous binary. After replacement,
failure leaves the new binary installed. Diagnose the calculator journal and
redeploy a reviewed working revision if needed. This workflow has not been run
against the production server as part of implementation.

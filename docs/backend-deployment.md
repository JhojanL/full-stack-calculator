# Backend deployment

The [backend deployment workflow](../.github/workflows/deploy-backend.yml) deploys
the Go API on pushes to `master` and also supports manual runs from `master`.

The workflow:

- audits the backend using the same project checks used by CI;
- builds the Linux AMD64 binary with `make -C backend build/api`;
- connects to the production droplet over verified SSH;
- atomically replaces `/srv/calculator/api`;
- restarts only `calculator.service`;
- verifies the local production healthcheck.

The workflow does not deploy the frontend, modify Caddy, provision the server, or
run database migrations.

## Production layout

The calculator currently uses:

| Resource | Value |
| --- | --- |
| Deployment user | `greenlight` |
| Runtime user | `calculator` |
| Application directory | `/srv/calculator` |
| Executable | `/srv/calculator/api` |
| systemd service | `calculator.service` |
| Internal API port | `4001` |
| Local healthcheck | `http://127.0.0.1:4001/healthcheck` |

The existing Greenlight API continues to use port `4000`. Port `4001` is reserved
for the calculator API and should not be exposed directly through the firewall.

The `greenlight` account is currently reused as the calculator deployment account
to keep the initial setup simple. A dedicated calculator deployment account can
be introduced later to further isolate deployment permissions.

## One-time server setup

The production server is an Ubuntu Linux x86_64 droplet. The deployment workflow
expects SSH, systemd, sudo, curl, jq, and standard GNU utilities to be available.

The server architecture and application port were verified with:

```bash
uname -m
sudo ss -lntp | grep ':4001' || true
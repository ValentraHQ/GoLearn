# Kubernetes deployment

```bash
kubectl apply -f namespace.yaml        # or: kubectl apply -k . (see below)
```

`kustomization.yaml` deploys the namespace, config, PostgreSQL (development only), the API and network policies. **It deliberately contains no Secrets.** Credentials are never committed to this repository: create them in the cluster (or sync them from a secret manager) *before* applying the rest, otherwise the pods stay in `CreateContainerConfigError` — the deployment fails closed rather than starting with a default password.

## 1. Create the secrets (out of band)

```bash
kubectl apply -f namespace.yaml

# One random password, used for both secrets. Never put it in a file in git.
PGPASS="$(openssl rand -hex 24)"

# Database credentials, read by the bundled PostgreSQL.
kubectl -n golearn create secret generic golearn-db \
  --from-literal=POSTGRES_USER=golearn \
  --from-literal=POSTGRES_PASSWORD="$PGPASS" \
  --from-literal=POSTGRES_DB=golearn

# Connection string, read by the API (envFrom). sslmode=require encrypts the connection.
kubectl -n golearn create secret generic golearn-secrets \
  --from-literal=DATABASE_URL="postgres://golearn:${PGPASS}@postgres.golearn.svc:5432/golearn?sslmode=require"

unset PGPASS
```

The password is hex, so it needs no URL escaping; if you choose your own, percent-encode it in `DATABASE_URL`.

Optional, only if you enable the remote runner (`GOLEARN_RUNNER=remote`, see `docs/sandbox.md`): add `GOLEARN_RUNNER_URL` to the ConfigMap and `GOLEARN_RUNNER_TOKEN` (≥16 random characters, the same value the runner daemon uses) to `golearn-secrets`:

```bash
kubectl -n golearn create secret generic golearn-secrets --dry-run=client -o yaml \
  --from-literal=DATABASE_URL="..." --from-literal=GOLEARN_RUNNER_TOKEN="$(openssl rand -hex 24)" \
  | kubectl apply -f -
```

For anything beyond a development cluster prefer a secret manager (External Secrets Operator, Sealed Secrets, your cloud's CSI driver) so the values never sit in shell history or manifests. `examples/secrets.example.yaml` shows the expected keys for those tools; it contains only placeholders, is not part of the kustomization, and must not be applied or filled in with real values in git.

## 2. Deploy

```bash
kubectl apply -k .
kubectl -n golearn rollout status deploy/golearn statefulset/postgres
```

## TLS to the database

The bundled single-node PostgreSQL generates a throw-away self-signed certificate in an init container and starts with `ssl=on`, so `sslmode=require` works out of the box: traffic is encrypted, but the server certificate is **not verified** (`require` does not check it). For production use a managed database or an operator with a real certificate and switch the URL to `sslmode=verify-full` (with the CA available to the API). The bundled database is for development clusters only.

## Code execution

`GOLEARN_RUNNER` is `disabled` in `config.yaml`. Enabling it is a separate hardening project (dedicated node pool, gVisor/Kata, an isolated runner daemon); see `docs/sandbox.md` and `docs/production-readiness.md`.

## Guard rails

`internal/deploycheck` (run by `go test ./...` and CI) fails the build if a Secret manifest, a placeholder password, or a non-TLS PostgreSQL URL is added to this directory, or if a secret the manifests reference is not documented here.

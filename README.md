# ship

Minimal infrastructure CLI for provisioning and controlling a single server on DigitalOcean, Hetzner, or Vultr.

## Requirements

- Go 1.23+
- Docker installed locally
- A `Dockerfile` in the current project
- One provider token exported in the environment:
  - `DIGITALOCEAN_TOKEN`
  - `HCLOUD_TOKEN`
  - `VULTR_API_KEY`
- SSH keys already registered with the selected provider

## Build

```bash
go build -o ship
```

## Usage

```bash
export DIGITALOCEAN_TOKEN=...

./ship server create
./ship deploy
./ship logs
./ship server destroy
```

Provider selection:

```bash
./ship server create --provider digitalocean
./ship server create --provider hetzner
./ship server create --provider vultr
```

## Commands

### `ship server create`

Creates a server with provider-specific defaults:

- DigitalOcean
  - Region: `nyc3`
  - Size: `s-2vcpu-4gb`
  - Image: `ubuntu-22-04-x64`
- Hetzner
  - Region: `nbg1`
  - Size: `cx22`
  - Image: `ubuntu-22.04`
- Vultr
  - Region: `ewr`
  - Size: `vc2-2c-4gb`
  - Image: `Ubuntu 22.04 x64`

The command waits for the server to become active, waits for SSH access, installs Docker, then stores server metadata in `.ship/server.json`.

Example output:

```text
STATUS=SERVER_CREATED
SERVER_ID=12345
SERVER_IP=1.2.3.4
```

Override defaults if needed:

```bash
./ship server create --provider digitalocean --region sfo3 --size s-1vcpu-2gb --image ubuntu-22-04-x64
./ship server create --provider hetzner --region fsn1 --size cpx21 --image ubuntu-24.04
./ship server create --provider vultr --region ord --size vc2-1c-2gb --image "Ubuntu 24.04 x64"
```

### `ship deploy`

Builds a local Docker image named `app`, saves it to `app.tar`, uploads it to the server, then runs:

```bash
docker load -i /root/app.tar
docker stop app || true
docker rm app || true
docker run -d --name app -p 80:80 app
```

Example output:

```text
STATUS=DEPLOY_COMPLETE
SERVER_IP=1.2.3.4
```

### `ship logs`

Fetches the last 100 log lines from the `app` container:

```bash
./ship logs
```

### `ship server destroy`

Deletes the server identified in `.ship/server.json` using the recorded provider and removes that local state file.

Example output:

```text
STATUS=SERVER_DESTROYED
```

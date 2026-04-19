<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-dark-transparent.png" />
    <img src="assets/logo.png" alt="ship Logo" width="320" />
  </picture>
</p>

<h1 align="center">ship</h1>

<p align="center"><strong>Infrastructure for AI Coding Agents</strong></p>

<p align="center">
An extremely lightweight infrastructure CLI for provisioning, deploying, tailing logs, and destroying servers.<br/>
One binary. Zero dashboards. A minimal cloud control layer that agents can drive reliably.
</p>

<p align="center">
  <a href="#install">Install</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#agent-skill">Agent Skill</a> &bull;
  <a href="https://github.com/basilysf1709/ship">GitHub</a> &bull;
  <a href="https://github.com/basilysf1709/ship/releases">Releases</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/providers-DigitalOcean%20%7C%20Hetzner%20%7C%20Vultr-blue" alt="providers" />
  <img src="https://img.shields.io/badge/license-MIT-green" alt="license" />
  <img src="https://img.shields.io/github/v/release/basilysf1709/ship?color=orange&label=version" alt="version" />
  <img src="https://img.shields.io/badge/language-Go-00ADD8" alt="language" />
</p>

> **Personal fork** — using this to learn Go CLI tooling and experiment with Hetzner deployments.

## What is ship?

`ship` is a minimal infrastructure primitive for AI coding agents.

There are too many moments where an agent is working inside the terminal, then suddenly has to break context to provision a server, deploy code, fetch logs, or tear infrastructure down. That context switch is wasteful. `ship` keeps the entire flow inside the CLI so deployment can be handled directly by cloud-capable agents without leaving the terminal.

The goal is simple: give agents a tiny, deterministic interface for infrastructure operations.

**How it works:**

1. **Create** a server with `ship server create`
2. **Deploy** the current project with `ship deploy`
3. **Inspect and operate** with `ship status`, `ship logs`, and `ship exec`
4. **Manage secrets and releases** with `ship secrets`, `ship release list`, and `ship rollback`
5. **List** locally tracked servers with `ship server list`
6. **Destroy** the server with `ship server destroy`

**Key features:**

- **Operator commands**: status, exec, secrets, release history, rollback, bootstrap, and domain setup
- **Single binary**: build once with `go build -o ship`
- **Provider support**: DigitalOcean, Hetzner, and Vultr
- **Deterministic output**: machine-friendly `KEY=VALUE` responses
- **Structured JSON mode**: pass `--json` for machine-readable output
- **No dashboard required**: everything happens from the terminal
- **Configurable deploy flow**: use `ship.json` for project-specific deploy steps
- **Local state tracking**: server metadata stored in `.ship/server.json`

## Build

```bash
go build -o ship
```

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/basilysf1709/ship/main/install.sh | sh
```

## Agent Skill

Download the reusable skill file dir

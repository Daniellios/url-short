## HTTPS with Certbot (Docker Compose)

This project uses:

- `nginx` on **80** (ACME + redirect) and **443** (TLS)
- Certificates in `./certbot/conf` on the host (mounted as `/etc/letsencrypt` in containers)
- A `certbot` service that runs `certbot renew` periodically

Config expects the certificate line to live at:

`certbot/conf/live/utilitools.tech/fullchain.pem`

(issue Certbot with **`-d utilitools.tech` first** so that folder name matches `nginx/nginx.conf`)

---

### 1) Prepare environment

In repo root `.env`:

- `LETSENCRYPT_EMAIL=you@example.com`
- `CERTBOT_DOMAINS=utilitools.tech www.utilitools.tech`
- `COMPOSE_DOMAIN=https://utilitools.tech` (optional but recommended)

DNS `A` records for `utilitools.tech` and `www` must point at the VPS. Open inbound **80** and **443**.

---

### 2) Chicken-and-egg: nginx needs certs, but webroot needs nginx

If `nginx` logs show:

`cannot load certificate "/etc/letsencrypt/live/utilitools.tech/fullchain.pem"`

then **webroot issuance is blocked** until files exist. Use **standalone** once (port 80 must be free).

From repo root:

```bash
docker compose stop nginx
docker compose run --rm -p 80:80 --entrypoint certbot certbot certonly --standalone \
  --email "$LETSENCRYPT_EMAIL" --agree-tos --no-eff-email \
  -d utilitools.tech -d www.utilitools.tech
docker compose up -d
```

If you do not use `.env` for the email, pass it explicitly:

```bash
docker compose run --rm -p 80:80 --entrypoint certbot certbot certonly --standalone \
  --email you@example.com --agree-tos --no-eff-email \
  -d utilitools.tech -d www.utilitools.tech
```

Then confirm on the host:

```bash
ls -la certbot/conf/live/utilitools.tech/
```

You should see `fullchain.pem` and `privkey.pem`.

---

### 3) When nginx already runs (renewal or new cert via webroot)

Start nginx (and dependencies) so `/.well-known/acme-challenge/` is reachable on port **80**:

```bash
docker compose up -d nginx
```

Issue:

```bash
docker compose run --rm --entrypoint certbot certbot certonly --webroot -w /var/www/certbot \
  --email "$LETSENCRYPT_EMAIL" --agree-tos --no-eff-email \
  -d utilitools.tech -d www.utilitools.tech
```

Reload:

```bash
docker compose exec nginx nginx -t && docker compose exec nginx nginx -s reload
```

---

### 4) Wrong directory under `live/`

If Certbot created e.g. `certbot/conf/live/www.utilitools.tech/` (because `www` was listed first), either:

- re-issue with **`-d utilitools.tech` before `-d www.utilitools.tech`**, or  
- change `ssl_certificate` paths in `nginx/nginx.conf` to match the folder you actually have.

---

### 5) Full stack

```bash
docker compose up -d --build
```

The long-running `certbot` container only **renews**; the first certificate must be obtained with `certonly` as above.

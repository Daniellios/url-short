## HTTPS with Certbot (Docker Compose)

This project is configured for:

- `nginx` serving HTTP (`80`) + HTTPS (`443`)
- `certbot` renewing certificates in the background
- ACME challenge path at `/.well-known/acme-challenge/`

### 1) Prepare environment

In repo root `.env`:

- `LETSENCRYPT_EMAIL=you@example.com`
- `CERTBOT_DOMAINS=utilitools.tech www.utilitools.tech`
- `COMPOSE_DOMAIN=https://utilitools.tech` (optional but recommended)

Ensure DNS `A` records point those domains to your server and ports `80/443` are open.

### 2) Start nginx first (HTTP challenge endpoint)

```bash
docker compose up -d nginx
```

### 3) Issue the first certificate

Run this from repo root:

```bash
docker compose run --rm certbot certonly --webroot -w /var/www/certbot \
  --email "$LETSENCRYPT_EMAIL" --agree-tos --no-eff-email \
  $(for d in $CERTBOT_DOMAINS; do printf -- "-d %s " "$d"; done)
```

On PowerShell, you can pass domains explicitly if needed:

```powershell
docker compose run --rm certbot certonly --webroot -w /var/www/certbot --email you@example.com --agree-tos --no-eff-email -d utilitools.tech -d www.utilitools.tech
```

### 4) Reload nginx to pick up the new cert

```bash
docker compose exec nginx nginx -s reload
```

### 5) Start full stack

```bash
docker compose up -d
```

`certbot` runs `certbot renew` every 12 hours in the compose service.

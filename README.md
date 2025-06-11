# Dynamic DNS Updater for Cloudflare

This Go-based containerized utility automatically updates your Cloudflare DNS A and AAAA records when your public IP address changes. Perfect for homelabs, media servers, Minecraft servers, or anything you host behind a dynamic IP.

---

## Features

* IPv4 and IPv6 support (A + AAAA records)
* Zone ID auto-discovery using your domain name
* Exponential backoff for IP lookup failures
* Detailed logging with timestamps
* Environment-variable driven configuration

---

## Configuration

Set the following environment variables when running the container or binary:

| Variable                 | Required | Default                   | Description                                                                       |
| ------------------------ | -------- | ------------------------- | --------------------------------------------------------------------------------- |
| `CF_API_TOKEN`           | ✅        | N/A                       | Cloudflare API token with DNS\:Edit and Zone\:Read permissions                    |
| `CF_DOMAIN`              | ✅        | N/A                       | Your base domain (e.g., `example.com`)                                            |
| `CF_RECORDS`             | ✅        | N/A                       | Comma-separated list of subdomain records to update (e.g., `home,minecraft,plex`) |
| `CHECK_INTERVAL_MINUTES` | ❌        | `5`                       | How often to check for IP changes                                                 |
| `IP_FILE`                | ❌        | `/tmp/last_ip`            | File to store last known IPs                                                      |
| `IP_CHECK_URL_V4`        | ❌        | `https://api.ipify.org`   | Service to retrieve public IPv4 address                                           |
| `IP_CHECK_URL_V6`        | ❌        | `https://api64.ipify.org` | Service to retrieve public IPv6 address                                           |

---

## Example Usage

```bash
docker run -d --name dyn-dns \
  -e CF_API_TOKEN=your_token \
  -e CF_DOMAIN=netlobo.com \
  -e CF_RECORDS=home,minecraft \
  -e CHECK_INTERVAL_MINUTES=3 \
  -v /data/dns:/data \
  -e IP_FILE=/data/last_ip \
  your-docker-image-name
```

---

## Notes

* You must pre-create the DNS A/AAAA records in Cloudflare.
* Records must live under the same `CF_DOMAIN` zone.
* This tool does not attempt to create records, only update them.

---

## Roadmap Ideas

* Add metrics (Prometheus-compatible?)
* Alerting on repeated failures
* Optional creation of missing DNS records
* Optional Slack/email notifications

---

MIT Licensed. Built for homelab reliability. ☁️

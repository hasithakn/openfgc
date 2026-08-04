
## Run the Demo

The full demo stack (consent portal, insurance portal demo app, event notifications, and a
provisioned WSO2 Identity Server tenant) runs entirely in Docker — **the only prerequisite is
Docker itself**, no Go/Node/Maven/etc. required on your machine.

```bash
./demo.sh start   # builds everything from source on first run, then starts the stack
```

One-time setup: add `127.0.0.1  wso2is` to `/etc/hosts` (the script will tell you if it's
missing and give you the exact command). Then open **http://localhost:3020**.

```echo '127.0.0.1  wso2is' | sudo tee -a /etc/hosts```

Other commands:

```bash
./demo.sh build   # (re)build all images from source — only needed after pulling code changes
./demo.sh stop    # stop containers, keep built images (fast restart with `start` again)
./demo.sh clean   # remove containers, volumes, images, and build output — start fresh
```

### Demo at a glance

| | URL |
|---|---|
| Insurance portal (start here) | http://localhost:3020 |
| Consent portal | http://localhost:5173 |
| WSO2 IS console (tenant: `insurance.org`) | https://wso2is:9443/carbon |
| OpenFGC API | http://localhost:8060 |
| Webhook callback URL | `http://webhook-listener:9091/webhook` |

| Login | Username | Password | Role |
|---|---|---|---|
| WSO2 IS tenant admin | `admin` | `admin` | full IS console access |
| Portal admin | `admin@gmail.com` | `Abc@1234` | `portaladmin` — all `portal:*` scopes |
| DPO | `dpo@gmail.com` | `Abc@1234` | `dpo` — grievances + own-consent scopes |

Self-registration is enabled (username + password only, auto-login on signup — no email/OTP
step). Tail webhook deliveries with `docker compose logs -f webhook-listener`.


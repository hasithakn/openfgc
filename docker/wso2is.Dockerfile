# Patches the vendor image so IS's self-referential URLs (issuer, /oauth2/authorize, JWKS,
# SCIM base) resolve to "wso2is" instead of "localhost" — Docker's embedded DNS already
# resolves that name for every other container; the browser needs exactly one one-time
# /etc/hosts entry (127.0.0.1 wso2is), printed by `demo.sh start` if missing.
#
# NOTE: this sed patch is best-effort against wso2/wso2is:7.3.0's default deployment.toml
# layout and has not yet been verified against a live container (see task: "Iterate
# prerequisites.sh against live wso2is:7.3.0 container" — this hostname patch is part of the
# same verify-against-the-real-image work). If the image's deployment.toml uses a different
# quoting style or key path, adjust the sed pattern below accordingly.
FROM wso2/wso2is:7.3.0

USER root
RUN set -eux; \
    conf_file=$(find / -maxdepth 6 -path '*/repository/conf/deployment.toml' 2>/dev/null | head -n1); \
    test -n "$conf_file"; \
    sed -i 's/^hostname = "localhost"/hostname = "wso2is"/' "$conf_file"; \
    grep -q 'hostname = "wso2is"' "$conf_file"
USER wso2carbon

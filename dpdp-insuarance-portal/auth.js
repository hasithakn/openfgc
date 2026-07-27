// Real WSO2 IS login for the insurance portal demo (authorization_code + PKCE).
// This module is this demo's whole "BFF" for auth: it owns the OIDC dance and the
// server-side session, and exposes a `requireAuth` middleware other routes use to
// resolve the caller's own identity — never trust a client-supplied userId again.
const crypto = require('crypto');
const cookieParser = require('cookie-parser');
const fetch = require('node-fetch');
const { Issuer, generators } = require('openid-client');
const config = require('./config.json');

const IS_ISSUER_URL = process.env.IS_ISSUER_URL || config.isIssuerUrl || 'https://localhost:9443/oauth2/token';
// /scim2/Me lives on the IS host itself, not under /oauth2/token.
const IS_SCIM_BASE_URL = IS_ISSUER_URL.replace(/\/oauth2\/token\/?$/, '');
const IS_CLIENT_ID = process.env.IS_CLIENT_ID || config.isClientId || '';
const IS_CLIENT_SECRET = process.env.IS_CLIENT_SECRET || config.isClientSecret || '';
const IS_REDIRECT_URI = process.env.IS_REDIRECT_URI || config.isRedirectUri || 'http://localhost:3020/auth/callback';
const SESSION_COOKIE_SECRET = process.env.SESSION_COOKIE_SECRET || config.sessionCookieSecret || 'dev-secret';
const PORT = process.env.PORT || config.port || 3020;
const POST_LOGOUT_REDIRECT_URI = `http://localhost:${PORT}/home.html`;
// Full SCIM PATCH path for the custom "age" claim, e.g. "urn:scim:wso2:schema:age" —
// only exists once a matching local claim has been created and SCIM2-mapped in the IS
// Console (see README). Leave empty in config.json to skip pushing age entirely.
const SCIM_AGE_ATTRIBUTE_PATH = process.env.SCIM_AGE_ATTRIBUTE_PATH || config.scimAgeAttributePath || '';

// Local WSO2 IS dev instances almost always run with a self-signed cert. Only relax
// TLS verification for discovery/token calls when the issuer itself is localhost —
// never do this against a real deployment.
if (/^https:\/\/(localhost|127\.0\.0\.1)/i.test(IS_ISSUER_URL)) {
  process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
  console.warn('[Auth] Relaxing TLS certificate verification for localhost IS discovery (dev only — do not do this against a real deployment).');
}

let clientPromise = null;
function getClient() {
  if (!clientPromise) {
    clientPromise = Issuer.discover(IS_ISSUER_URL).then((issuer) => new issuer.Client({
      client_id: IS_CLIENT_ID,
      client_secret: IS_CLIENT_SECRET,
      redirect_uris: [IS_REDIRECT_URI],
      response_types: ['code'],
    }));
  }
  return clientPromise;
}

// ── Session store (in-memory; fine for a single-process demo, not for production) ──
const sessions = new Map(); // sessionId -> { sub, email, name, idToken, expiresAt }
const SESSION_TTL_MS = 8 * 60 * 60 * 1000; // 8 hours

function createSession(claims, idToken) {
  const sessionId = crypto.randomBytes(24).toString('hex');
  const email = claims.email || claims.sub;
  sessions.set(sessionId, {
    sub: claims.sub,
    email,
    name: claims.name || claims.given_name || (email || '').split('@')[0],
    idToken,
    expiresAt: Date.now() + SESSION_TTL_MS,
  });
  return sessionId;
}

function getSession(sessionId) {
  if (!sessionId) return null;
  const session = sessions.get(sessionId);
  if (!session) return null;
  if (session.expiresAt < Date.now()) {
    sessions.delete(sessionId);
    return null;
  }
  return session;
}

const SESSION_COOKIE = 'insurance_session';
const TXN_COOKIE = 'insurance_auth_txn';
const PENDING_PII_COOKIE = 'insurance_pending_pii';
const COOKIE_OPTS = { httpOnly: true, sameSite: 'lax', signed: true, secure: false }; // secure:false — this demo runs over plain http locally

// Best-effort: if this login's email matches PII stashed by the quotation form before
// any account existed, push the applicant's name (and age, if a custom SCIM claim for
// it has been configured) into their IS profile via SCIM2 — so a customer never has
// to retype what they already gave us.
async function applyPendingPii(req, res, email, accessToken) {
  const raw = req.signedCookies[PENDING_PII_COOKIE];
  res.clearCookie(PENDING_PII_COOKIE);
  if (!raw) {
    console.log('[Auth] No pending-PII cookie present for this login (' + email + ') — nothing to push to SCIM.');
    return;
  }

  let pending;
  try {
    pending = JSON.parse(raw);
  } catch (e) {
    console.warn('[Auth] Pending-PII cookie was present but not valid JSON — skipping SCIM push.');
    return;
  }
  // Only ever apply pending PII to the identity it was actually collected for — the
  // real trust boundary is here, not at collection time (the cookie itself is set
  // before any login exists, so it can't be verified against an identity until now).
  if (!pending.email || !email || pending.email.toLowerCase() !== email.toLowerCase()) {
    console.log('[Auth] Pending-PII email (' + pending.email + ') does not match logged-in email (' + email + ') — skipping SCIM push.');
    return;
  }

  const parts = String(pending.name || '').trim().split(/\s+/).filter(Boolean);
  const givenName = parts[0] || '';
  const familyName = parts.slice(1).join(' ');

  // Sent as two independent PATCH calls (not one combined request) — SCIM PATCH is
  // atomic, so a failing custom "age" attribute (e.g. not yet wired into IS's SCIM2
  // schema) must never be able to roll back the perfectly valid name update too.
  if (givenName) {
    await scimPatchMe(accessToken, email, 'name', [
      { op: 'replace', path: 'name', value: { givenName, familyName } },
    ]);
  }

  const ageNumber = Number(pending.age);
  if (SCIM_AGE_ATTRIBUTE_PATH && pending.age && Number.isFinite(ageNumber)) {
    // The custom "age" claim is configured as an Integer in IS — sending it as a JSON
    // string (rather than a real number) trips IS's "DataType doesn't match" check.
    await scimPatchMe(accessToken, email, 'age', [
      { op: 'replace', path: SCIM_AGE_ATTRIBUTE_PATH, value: ageNumber },
    ]);
  }

  if (!givenName && !(SCIM_AGE_ATTRIBUTE_PATH && pending.age)) {
    console.log('[Auth] Pending-PII matched ' + email + ' but had nothing usable to push (no name, no age/no age attribute configured).');
  }
}

async function scimPatchMe(accessToken, email, label, operations) {
  console.log(`[Auth] Pushing SCIM ${label} update for ${email}:`, JSON.stringify(operations));
  try {
    const r = await fetch(`${IS_SCIM_BASE_URL}/scim2/Me`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/scim+json',
      },
      body: JSON.stringify({
        schemas: ['urn:ietf:params:scim:api:messages:2.0:PatchOp'],
        Operations: operations,
      }),
    });
    if (!r.ok) {
      console.warn(`[Auth] SCIM ${label} update failed:`, r.status, await r.text().catch(() => ''));
    } else {
      console.log(`[Auth] SCIM ${label} updated from quotation PII for`, email);
    }
  } catch (e) {
    console.warn(`[Auth] SCIM ${label} update error:`, e.message);
  }
}

function requireAuth(req, res, next) {
  const session = getSession(req.signedCookies[SESSION_COOKIE]);
  if (!session) {
    if (req.path.startsWith('/api/')) {
      return res.status(401).json({ error: 'not authenticated' });
    }
    return res.redirect('/auth/login');
  }
  req.user = session;
  next();
}

function register(app) {
  app.use(cookieParser(SESSION_COOKIE_SECRET));

  // Both "Log In" and "Create Account" point here — WSO2 IS's own hosted login page
  // offers a "Register" link once self-registration is enabled for the tenant, so a
  // single flow covers both without a separate onboarding page.
  app.get('/auth/login', async (req, res) => {
    if (!IS_CLIENT_ID || !IS_CLIENT_SECRET) {
      return res.status(500).send('WSO2 IS is not configured yet — set isClientId/isClientSecret in config.json.');
    }
    try {
      const client = await getClient();
      const codeVerifier = generators.codeVerifier();
      const codeChallenge = generators.codeChallenge(codeVerifier);
      const state = generators.state();

      res.cookie(TXN_COOKIE, JSON.stringify({ codeVerifier, state }), { ...COOKIE_OPTS, maxAge: 5 * 60 * 1000 });

      const url = client.authorizationUrl({
        // internal_login is required for the SCIM2 /scim2/Me call in applyPendingPii
        // below — this app must be authorized for that scope in the IS Console (same
        // requirement as the Consent Portal's own profile page).
        scope: 'openid email profile internal_login',
        state,
        code_challenge: codeChallenge,
        code_challenge_method: 'S256',
      });
      res.redirect(url);
    } catch (e) {
      console.error('[Auth] login error:', e.message);
      res.status(500).send('Unable to reach WSO2 IS. Check isIssuerUrl in config.json and that IS is running.');
    }
  });

  app.get('/auth/callback', async (req, res) => {
    try {
      const txnRaw = req.signedCookies[TXN_COOKIE];
      if (!txnRaw) return res.status(400).send('Login session expired — please try again from /home.html.');
      const { codeVerifier, state } = JSON.parse(txnRaw);
      res.clearCookie(TXN_COOKIE);

      const client = await getClient();
      const params = client.callbackParams(req);
      const tokenSet = await client.callback(IS_REDIRECT_URI, params, { code_verifier: codeVerifier, state });
      const claims = tokenSet.claims();

      const sessionId = createSession(claims, tokenSet.id_token);
      const session = getSession(sessionId);
      res.cookie(SESSION_COOKIE, sessionId, { ...COOKIE_OPTS, maxAge: SESSION_TTL_MS });

      // Use the session's already-resolved email (claims.email falls back to claims.sub
      // there) — IS isn't releasing the `email` claim itself for every login, but sub
      // is always present and equals the username, which is the email for this org.
      await applyPendingPii(req, res, session.email, tokenSet.access_token);

      res.redirect('/account.html');
    } catch (e) {
      console.error('[Auth] callback error:', e.message);
      res.status(500).send('Login failed: ' + e.message);
    }
  });

  app.get('/auth/logout', async (req, res) => {
    const sessionId = req.signedCookies[SESSION_COOKIE];
    const session = getSession(sessionId);
    sessions.delete(sessionId);
    res.clearCookie(SESSION_COOKIE);
    try {
      const client = await getClient();
      if (client.issuer.end_session_endpoint) {
        const url = client.endSessionUrl({
          id_token_hint: session ? session.idToken : undefined,
          post_logout_redirect_uri: POST_LOGOUT_REDIRECT_URI,
        });
        return res.redirect(url);
      }
    } catch (e) {
      // Fall through to a plain local redirect — the local session is already cleared.
    }
    res.redirect('/home.html');
  });

  app.get('/api/me', requireAuth, (req, res) => {
    res.json({ email: req.user.email, name: req.user.name });
  });

  // Called by the quotation form right after a consent is created, before the user
  // has any account yet — stashes the name/age they just typed in so they can be
  // pushed into their IS profile via SCIM the moment they log in with that same email.
  app.post('/api/pending-pii', (req, res) => {
    const { email, name, age } = req.body || {};
    if (!email || !name) return res.status(400).json({ error: 'email and name are required' });
    res.cookie(PENDING_PII_COOKIE, JSON.stringify({ email, name, age }), { ...COOKIE_OPTS, maxAge: 30 * 60 * 1000 });
    res.json({ ok: true });
  });
}

module.exports = { register, requireAuth };

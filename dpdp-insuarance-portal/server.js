const express = require('express');
const fetch = require('node-fetch');
const path = require('path');
const config = require('./config.json');
const auth = require('./auth');

const app = express();
app.use(express.json());
auth.register(app);

// Gate account.html / learnpath-account.html themselves (not just their API calls) — this
// must come before the static middleware below so an unauthenticated request never even
// gets the file.
app.get('/account.html', auth.requireAuth, (_, res) => {
  res.sendFile(path.join(__dirname, 'public', 'account.html'));
});
app.get('/learnpath-account.html', auth.requireAuth, (_, res) => {
  res.sendFile(path.join(__dirname, 'public', 'learnpath-account.html'));
});

app.use(express.static(path.join(__dirname, 'public')));

// Clean URL aliases (no .html extension needed)
app.get('/delegate',  (_, res) => res.redirect('/delegate.html'));
app.get('/quotation', (_, res) => res.redirect('/quotation.html'));
app.get('/home',      (_, res) => res.redirect('/home.html'));

const OPENFGC_BASE       = process.env.OPENFGC_BASE        || config.openfgcUrl        || 'http://localhost:3000';
const ORG_ID             = process.env.ORG_ID              || config.orgId             || 'insurance-org';
const PORT               = process.env.PORT                || config.port              || 3020;
const PORTAL_BACKEND_URL = process.env.PORTAL_BACKEND_URL  || config.portalBackendUrl  || 'http://localhost:3001';
const CONSENT_PORTAL_URL = process.env.CONSENT_PORTAL_URL  || config.consentPortalUrl  || 'http://localhost:5173';

// ─── API Request Logger ───────────────────────────────────────────────────────
const apiLogs = [];

function logCall(method, url, reqBody, status, resBody) {
  apiLogs.unshift({
    id: Date.now() + '_' + Math.random().toString(36).slice(2, 6),
    ts: new Date().toISOString(),
    method,
    url,
    reqBody: reqBody || null,
    status,
    resBody
  });
  if (apiLogs.length > 40) apiLogs.pop();
}

async function fgcFetch(method, urlPath, body) {
  const url = `${OPENFGC_BASE}${urlPath}`;
  const opts = {
    method,
    headers: {
      'Content-Type': 'application/json',
      'org-id': ORG_ID,
      'group-id': ORG_ID
    }
  };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const r = await fetch(url, opts);
  const text = await r.text();
  let data;
  try { data = JSON.parse(text); } catch (e) { data = { raw: text }; }
  logCall(method, url, body !== undefined ? body : null, r.status, data);
  return { ok: r.ok, status: r.status, data };
}

// ─── Routes ───────────────────────────────────────────────────────────────────

app.get('/api/logs', (_, res) => {
  res.json(apiLogs.slice(0, 20));
});

app.get('/api/config', (_, res) => {
  res.json({ companyName: config.companyName, orgId: ORG_ID, consentPortalUrl: CONSENT_PORTAL_URL, enableMinorFlow: config.enableMinorFlow !== false });
});

// Store PII for a user in the one-stop portal backend
app.post('/api/store-user-pii', async (req, res) => {
  try {
    const { userId, pii } = req.body;
    if (!userId || !Array.isArray(pii)) return res.status(400).json({ error: 'userId and pii[] required' });
    // Ensure user record exists (409 is fine — already there)
    await fetch(`${PORTAL_BACKEND_URL}/api/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: userId, email: userId })
    });
    const r = await fetch(`${PORTAL_BACKEND_URL}/api/users/${encodeURIComponent(userId)}/pii`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(pii)
    });
    const text = await r.text();
    let data;
    try { data = JSON.parse(text); } catch (e) { data = {}; }
    logCall('PUT', `${PORTAL_BACKEND_URL}/api/users/${userId}/pii`, pii, r.status, data);
    res.status(r.status).json(data);
  } catch (e) {
    console.error('[PII] Error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

app.get('/api/elements', async (_, res) => {
  try {
    const { status, data } = await fgcFetch('GET', '/api/v1/consent-elements?limit=100');
    res.status(status).json(data);
  } catch (e) {
    console.error('[Elements] GET error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

app.get('/api/purposes', async (req, res) => {
  try {
    const qs = req.query.name
      ? `?name=${encodeURIComponent(req.query.name)}`
      : '?limit=50';
    const { status, data } = await fgcFetch('GET', `/api/v1/consent-purposes${qs}`);
    res.status(status).json(data);
  } catch (e) {
    console.error('[Purposes] GET error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

// List consents for the AUTHENTICATED caller only. userId comes from the server-side
// session (auth.requireAuth), never from client input — previously this trusted a
// client-supplied ?userId=, which let anyone view anyone else's consents by editing
// the URL.
app.get('/api/user-consents', auth.requireAuth, async (req, res) => {
  try {
    const userId = req.user.email;
    let qs = `?userIds=${encodeURIComponent(userId)}&limit=50`;
    if (req.query.purposeName) qs += `&purposeName=${encodeURIComponent(req.query.purposeName)}`;
    const { status, data } = await fgcFetch('GET', `/api/v1/consents${qs}`);
    res.status(status).json(data);
  } catch (e) {
    console.error('[UserConsents] GET error:', e.message);
    res.status(500).json({ error: e.message });
  }
});


// Get a single consent by ID
app.get('/api/consents/:id', async (req, res) => {
  try {
    const { status, data } = await fgcFetch('GET', `/api/v1/consents/${req.params.id}`);
    res.status(status).json(data);
  } catch (e) {
    console.error('[Consent] GET error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

// Revoke a consent
app.post('/api/consents/:id/revoke', async (req, res) => {
  try {
    const body = {
      actionBy: req.body.actionBy || 'user',
      revocationReason: req.body.revocationReason || 'User request'
    };
    const { status, data } = await fgcFetch('PUT', `/api/v1/consents/${req.params.id}/revoke`, body);
    res.status(status).json(data);
  } catch (e) {
    console.error('[Consent] Revoke error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

// Anonymize a consent — replace userId with anon token, update via PUT
app.post('/api/consents/:id/anonymize', async (req, res) => {
  try {
    const { id } = req.params;
    const anonId = 'anon_' + id.replace(/-/g, '').substring(0, 8);

    const getResult = await fgcFetch('GET', `/api/v1/consents/${id}`);
    if (!getResult.ok) return res.status(getResult.status).json(getResult.data);

    const c = getResult.data;
    const putPayload = {
      type: c.type,
      validityTime: c.validityTime,
      recurringIndicator: c.recurringIndicator || false,
      dataAccessValidityDuration: c.dataAccessValidityDuration || 0,
      frequency: c.frequency || 0,
      purposes: c.purposes,
      attributes: Object.assign({}, c.attributes, { userId: anonId }),
      authorizations: (c.authorizations || []).map(a => {
        const s = (a.status || '').toUpperCase();
        const safeStatus = s.startsWith('SYS_') ? 'REJECTED' : a.status;
        return Object.assign({}, a, { userId: anonId, status: safeStatus });
      })
    };

    const putResult = await fgcFetch('PUT', `/api/v1/consents/${id}`, putPayload);
    res.status(putResult.status).json(putResult.data);
  } catch (e) {
    console.error('[Consent] Anonymize error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

// Clear API logs
app.delete('/api/logs', (_, res) => {
  apiLogs.length = 0;
  res.json({ cleared: true });
});

// Create consent record after quotation form submission
app.post('/api/consents', async (req, res) => {
  try {
    const { userId, purposes, validityTime, expirationTime, authorizations, type } = req.body;
    const expiry = expirationTime || validityTime || (Date.now() + (90 * 24 * 60 * 60 * 1000));
    const payload = {
      type: type || 'insurance',
      expirationTime: expiry,
      recurringIndicator: false,
      dataAccessValidityDuration: 0,
      frequency: 0,
      purposes,
      authorizations: authorizations || [
        { userId: userId || 'anonymous', type: 'authorisation', status: 'APPROVED' }
      ]
    };
    console.log(`[Consent] Creating for userId=${userId}, purposes=${purposes.map(p => p.name).join(', ')}`);
    const { ok, status, data } = await fgcFetch('POST', '/api/v1/consents', payload);
    if (ok) {
      console.log(`[Consent] Created — id=${data.id}`);
    } else {
      console.error(`[Consent] Create failed (HTTP ${status}):`, data);
    }
    res.status(status).json(data);
  } catch (e) {
    console.error('[Consent] Error:', e.message);
    res.status(500).json({ error: e.message });
  }
});

app.listen(PORT, () => {
  console.log(`\nLife Insurance Consent Demo`);
  console.log(`  Entry      → http://localhost:${PORT}/`);
  console.log(`  Home       → http://localhost:${PORT}/home.html`);
  console.log(`  Quotation  → http://localhost:${PORT}/quotation.html`);
  console.log(`  Delegate   → http://localhost:${PORT}/delegate`);
  console.log(`  Account    → http://localhost:${PORT}/account.html`);
  console.log(`\n  Org ID     : ${ORG_ID}`);
  console.log(`  OpenFGC    : ${OPENFGC_BASE}\n`);
});

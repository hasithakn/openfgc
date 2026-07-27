const http = require('http');

const PORT = process.env.PORT || 9091;
const DIVIDER = '─'.repeat(50);

// Only these carry real information about the delivery — the rest (connection, host,
// content-length, user-agent, upgrade, etc.) is HTTP/transport plumbing, not signal.
const EVENT_HEADER_FIELDS = [
    ['org-id', 'x-org-id'],
    ['event-id', 'x-event-id'],
    ['delivery-id', 'x-delivery-id'],
    ['signature', 'x-event-signature'],
];

function logRequest(method, path, headers) {
    console.log(`\n${DIVIDER}`);
    console.log(`[${new Date().toISOString()}] ${method} ${path}`);
    for (const [label, header] of EVENT_HEADER_FIELDS) {
        if (headers[header]) {
            console.log(`  ${label.padEnd(11)}: ${headers[header]}`);
        }
    }
}

function logPayload(body) {
    try {
        const indented = JSON.stringify(JSON.parse(body), null, 2)
            .split('\n')
            .map(line => `  ${line}`)
            .join('\n');
        console.log(`  payload:\n${indented}`);
    } catch {
        console.log(`  payload (non-JSON): ${body}`);
    }
}

const server = http.createServer((req, res) => {
    const parsedUrl = new URL(req.url, `http://${req.headers.host || 'localhost'}`);
    const method = req.method;
    const path = parsedUrl.pathname;
    const query = Object.fromEntries(parsedUrl.searchParams);

    logRequest(method, path, req.headers);

    // GET Request - Challenge Verification (e.g. Webhook / WebSub intent verification)
    if (method === 'GET') {
        const challenge = query.challenge || query['hub.challenge'] || query.challenge_token || query.token;
        if (challenge) {
            console.log(`  -> echoing back challenge query parameter: "${challenge}"`);
            res.writeHead(200, { 'Content-Type': 'text/plain' });
            res.end(challenge);
        } else {
            console.log(`  -> no challenge parameter present`);
            res.writeHead(200, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({ status: 'active', message: 'Webhook listener running', query }));
        }
        return;
    }

    // POST / PUT Request - Event Delivery / Push Payload
    if (method === 'POST' || method === 'PUT') {
        let body = '';
        req.on('data', chunk => {
            body += chunk.toString();
        });

        req.on('end', () => {
            logPayload(body);

            try {
                const parsedBody = JSON.parse(body);
                // Check if challenge is inside JSON payload
                const challenge = parsedBody.challenge || parsedBody['hub.challenge'] || parsedBody.challenge_token;

                if (challenge) {
                    console.log(`  -> echoing back challenge from JSON body: "${challenge}"`);
                    res.writeHead(200, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ challenge: challenge, status: 'verified' }));
                } else {
                    res.writeHead(200, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ status: 'received', timestamp: new Date().toISOString() }));
                }
            } catch (e) {
                // Non-JSON or raw text body
                if (body && body.includes('challenge')) {
                    res.writeHead(200, { 'Content-Type': 'text/plain' });
                    res.end(body);
                } else {
                    res.writeHead(200, { 'Content-Type': 'text/plain' });
                    res.end('OK');
                }
            }
        });
        return;
    }

    res.writeHead(405, { 'Content-Type': 'text/plain' });
    res.end('Method Not Allowed');
});

server.listen(PORT, () => {
    console.log(DIVIDER);
    console.log(` Webhook Echo Listener started on http://localhost:${PORT}`);
    console.log(` Ready to echo challenge and receive Webhook events.`);
    console.log(` Callback URL to use in subscription: http://localhost:${PORT}/webhook`);
    console.log(DIVIDER);
});

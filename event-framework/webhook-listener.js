const http = require('http');
const url = require('url');

const PORT = process.env.PORT || 9091;

const server = http.createServer((req, res) => {
    const parsedUrl = url.parse(req.url, true);
    const method = req.method;
    const path = parsedUrl.pathname;
    const query = parsedUrl.query;

    const { 'http2-settings': _http2Settings, ...loggedHeaders } = req.headers;

    console.log(`\n==================================================`);
    console.log(`[${new Date().toISOString()}] Received ${method} request on ${path}`);
    console.log(`Headers:`, loggedHeaders);

    // GET Request - Challenge Verification (e.g. Webhook / WebSub intent verification)
    if (method === 'GET') {
        const challenge = query.challenge || query['hub.challenge'] || query.challenge_token || query.token;
        if (challenge) {
            console.log(`-> Echoing back challenge query parameter: "${challenge}"`);
            res.writeHead(200, { 'Content-Type': 'text/plain' });
            res.end(challenge);
        } else {
            console.log(`-> GET request received (no challenge parameter found).`);
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
            console.log(`-> Request Body:\n${body}`);

            try {
                const parsedBody = JSON.parse(body);
                // Check if challenge is inside JSON payload
                const challenge = parsedBody.challenge || parsedBody['hub.challenge'] || parsedBody.challenge_token;

                if (challenge) {
                    console.log(`-> Echoing back challenge from JSON body: "${challenge}"`);
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
    console.log(`==================================================`);
    console.log(` Webhook Echo Listener started on http://localhost:${PORT}`);
    console.log(` Ready to echo challenge and receive Webhook events.`);
    console.log(` Callback URL to use in subscription: http://localhost:${PORT}/webhook`);
    console.log(`==================================================`);
});

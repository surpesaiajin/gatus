# Mock local: emissor de token OAuth2 (client credentials) + API protegida.
import json, sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import parse_qs
import base64

TOKEN = "ptoss2-token"
class H(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/token":
            self.send_response(404); self.end_headers(); return
        body = parse_qs(self.rfile.read(int(self.headers.get("Content-Length", 0))).decode())
        auth = self.headers.get("Authorization", "")
        cid = csec = None
        if auth.startswith("Basic "):
            cid, csec = base64.b64decode(auth[6:]).decode().split(":", 1)
        cid = cid or body.get("client_id", [None])[0]
        csec = csec or body.get("client_secret", [None])[0]
        ok = body.get("grant_type") == ["client_credentials"] and cid == "gatus-client" and csec == "gatus-secret"
        self.send_response(200 if ok else 401)
        self.send_header("Content-Type", "application/json"); self.end_headers()
        self.wfile.write(json.dumps({"access_token": TOKEN, "token_type": "bearer", "expires_in": 3600} if ok else {"error": "invalid_client"}).encode())
    def do_GET(self):
        ok = self.path == "/api" and self.headers.get("Authorization") == "Bearer " + TOKEN
        self.send_response(200 if ok else 401); self.end_headers()
        self.wfile.write(b"ok" if ok else b"unauthorized")
    def log_message(self, fmt, *a):
        sys.stderr.write("[mock] " + (fmt % a) + " auth=" + str(self.headers.get("Authorization", "-"))[:20] + "\n")

HTTPServer(("127.0.0.1", int(sys.argv[1])), H).serve_forever()

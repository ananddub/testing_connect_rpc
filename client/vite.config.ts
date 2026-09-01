import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import basicSsl from '@vitejs/plugin-basic-ssl'
import http2 from 'node:http2'

// Custom Pure HTTP/2 (ALPN: h2) Proxy Middleware for ConnectRPC
function http2ConnectProxy(): Plugin {
  let session: http2.ClientHttp2Session | null = null;

  function getSession(): http2.ClientHttp2Session {
    if (!session || session.closed || session.destroyed) {
      session = http2.connect('https://localhost:8085', {
        rejectUnauthorized: false,
      });
      session.on('error', (err) => {
        console.error('[HTTP/2 Proxy Session Error]', err);
        session = null;
      });
    }
    return session;
  }

  return {
    name: 'http2-connect-proxy',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (!req.url?.startsWith('/todo.v1.TodoService/')) {
          return next();
        }

        const h2Session = getSession();
        const headers: http2.OutgoingHttpHeaders = {
          ':method': req.method,
          ':path': req.url,
          ...req.headers,
        };
        delete headers['host'];
        delete headers['connection'];
        delete headers['keep-alive'];
        delete headers['proxy-connection'];

        const h2Req = h2Session.request(headers);

        req.pipe(h2Req);

        h2Req.on('response', (responseHeaders) => {
          const status = Number(responseHeaders[':status'] || 200);
          const outHeaders: Record<string, string | string[]> = {};
          for (const [k, v] of Object.entries(responseHeaders)) {
            if (!k.startsWith(':') && v !== undefined) {
              outHeaders[k] = v as string | string[];
            }
          }
          res.writeHead(status, outHeaders);
        });

        h2Req.on('data', (chunk) => {
          res.write(chunk);
        });

        h2Req.on('end', () => {
          res.end();
        });

        h2Req.on('error', (err) => {
          console.error('[HTTP/2 Proxy Request Error]', err);
          if (!res.headersSent) {
            res.writeHead(502);
            res.end('Bad Gateway: ' + err.message);
          }
        });
      });
    },
  };
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    basicSsl(),
    http2ConnectProxy(),
  ],
  server: {
    port: 5173,
    host: true,
  },
})

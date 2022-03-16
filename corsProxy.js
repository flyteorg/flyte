const http = require('http');
const https = require('https');
const url = require('url');

/** Mounts a proxy at `basePath` such that requests of the form
 * `${basePath}/http://www.example.com?queryString=value` will be converted to
 * `http://www.example.com?queryString=value`, preserving any request headers.
 * The response is piped directly to the client with no processing.
 */
module.exports = function corsProxy(basePath) {
  const pathOffset = basePath.length;
  return function (req, res, next) {
    const pathIndex = req.url.indexOf(basePath);
    // base path doesn't match, ignore
    if (pathIndex < 0) {
      return next();
    }

    const targetUrl = req.url.substring(pathIndex + pathOffset + 1);
    const config = url.parse(targetUrl);
    config.method = req.method;

    // We need to use a different request class depending on the
    // protocol
    const handler = config.protocol === 'https:' ? https : http;

    const proxyReq = handler.request(config, function (proxyRes) {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res, {
        end: true,
      });
    });

    proxyReq.on('error', (err) => res.status(500).send(err));

    // Copy over all headers except for 'Host', since that value would
    // point to *this* server, and not the remote server.
    Object.keys(req.headers).forEach((key) => {
      if (key.toLowerCase() === 'host') {
        return;
      }
      proxyReq.setHeader(key, req.headers[key]);
    });

    req.pipe(proxyReq, {
      end: true,
    });
  };
};

/** @type {import('next').NextConfig} */
const backend =
  (process.env.BACKEND_PROXY_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');

const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ['lightweight-charts'],
  // Default 60s is too tight on slow machines / heavy prerender; avoids worker SIGTERM loops.
  staticPageGenerationTimeout: 180,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${backend}/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  allowedDevOrigins: ["localhost:3000", "localhost:3001", "0.0.0.0:3000", "0.0.0.0:3001", "192.168.0.177:3000", "192.168.0.177:3001"],
};

module.exports = nextConfig;

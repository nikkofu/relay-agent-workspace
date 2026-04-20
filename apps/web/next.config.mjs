/** @type {import('next').NextConfig} */
const nextConfig = {
  // Next.js 16 explicit cache model
  cacheComponents: true,
  // Enable React Compiler for performance (requires React 19+)
  reactCompiler: true,
  // Ensure images are configured if needed
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'github.com',
      },
      {
        protocol: 'https',
        hostname: 'avatars.githubusercontent.com',
      },
    ],
  },
};

export default nextConfig;

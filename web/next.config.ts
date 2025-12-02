import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  images: {
    // Image optimization disabled for deployment compatibility (Netlify/Vercel Free tier)
    // Avatar images are loaded via Radix UI Avatar component from dicebear API
    unoptimized: true,
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'api.dicebear.com',
      },
    ],
  },
};

export default nextConfig;

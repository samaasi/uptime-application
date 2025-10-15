import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactCompiler: true,
  images: {
    domains: ['placehold.co', 'github.com'],
  },
};

export default nextConfig;

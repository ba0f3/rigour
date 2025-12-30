import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  env: {
    // Workaround to inject env vars into the build time files, to be replaced at container runtime
    // See entrypoint.sh and Dockerfile for more details
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'NEXT_PUBLIC_API_BASE_URL_PLACEHOLDER',
  },
};

export default nextConfig;

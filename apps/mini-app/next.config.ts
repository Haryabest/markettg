import type { NextConfig } from "next";

const apiOrigin = process.env.API_URL || "http://localhost:8080";

const nextConfig: NextConfig = {
  output: "standalone",
  transpilePackages: ["@heroui/react"],
  images: {
    remotePatterns: [
      { protocol: "http", hostname: "localhost", port: "9000" },
      { protocol: "https", hostname: "**" },
    ],
  },
  async rewrites() {
    return [
      { source: "/api/:path*", destination: `${apiOrigin}/api/:path*` },
    ];
  },
};

export default nextConfig;

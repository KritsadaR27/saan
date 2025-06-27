/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  experimental: {
    appDir: true,
  },
  images: {
    domains: ['localhost'],
  },
  env: {
    NEXT_PUBLIC_CHAT_API_URL: process.env.NEXT_PUBLIC_CHAT_API_URL,
    NEXT_PUBLIC_ORDER_API_URL: process.env.NEXT_PUBLIC_ORDER_API_URL,
    NEXT_PUBLIC_INVENTORY_API_URL: process.env.NEXT_PUBLIC_INVENTORY_API_URL,
    NEXT_PUBLIC_DELIVERY_API_URL: process.env.NEXT_PUBLIC_DELIVERY_API_URL,
    NEXT_PUBLIC_FINANCE_API_URL: process.env.NEXT_PUBLIC_FINANCE_API_URL,
  }
}

module.exports = nextConfig

export const config = {
  baseUrl: process.env.BASE_URL || 'http://localhost:8081',
  businessType: process.env.BUSINESS_TYPE || 'professional',
  dbName: process.env.DB_NAME || 'service1',
  dbHost: process.env.DB_HOST || 'localhost',
  dbPort: parseInt(process.env.DB_PORT || '5432'),
  dbUser: process.env.DB_USER || 'postgres',
  dbPassword: process.env.DB_PASSWORD || 'postgres',
};

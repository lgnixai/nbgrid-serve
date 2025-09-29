/* Minimal smoke test using built dist */
const { Teable } = require('../dist/index.js');

async function main() {
  const teable = new Teable({ baseUrl: process.env.TEABLE_BASE_URL || 'http://localhost:3000', debug: true });
  try {
    console.log('Health check...');
    const health = await teable.healthCheck();
    console.log('Health:', health);
  } catch (e) {
    console.error('Health check failed (expected if backend not running):', e && e.message);
  }

  try {
    console.log('List spaces (unauth) should 401...');
    await teable.listSpaces({ limit: 1 });
  } catch (e) {
    console.log('List spaces error (ok):', (e && e.code) || e && e.message);
  }
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});


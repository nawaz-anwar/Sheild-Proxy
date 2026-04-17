const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

test('dashboard includes analytics and domain management sections', () => {
  const pagePath = path.join(__dirname, '..', 'pages', 'index.vue');
  const source = fs.readFileSync(pagePath, 'utf8');
  assert.match(source, /Domain management/);
  assert.match(source, /Traffic chart/);
  assert.match(source, /top IPs/i);
});

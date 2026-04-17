import { d as defineEventHandler } from '../../nitro/nitro.mjs';
import 'node:http';
import 'node:https';
import 'node:events';
import 'node:buffer';
import 'node:fs';
import 'node:path';
import 'node:crypto';
import 'node:url';

const metrics_get = defineEventHandler(() => "shield_dashboard_up 1\n");

export { metrics_get as default };
//# sourceMappingURL=metrics.get.mjs.map

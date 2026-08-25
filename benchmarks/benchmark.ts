import http from 'k6/http';
import {check} from 'k6';
import exec from 'k6/execution';
import {Options} from 'k6/options';
import {Trend, Counter} from 'k6/metrics';

// Rolling window metrics - Latency (0.1s, 0.5s, 1s)
const rollingAvgLatency01s = new Trend('rolling_avg_latency_0_1s', true);
const rollingP95Latency01s = new Trend('rolling_p95_latency_0_1s', true);
const rollingAvgLatency05s = new Trend('rolling_avg_latency_0_5s', true);
const rollingP95Latency05s = new Trend('rolling_p95_latency_0_5s', true);
const rollingAvgLatency1s = new Trend('rolling_avg_latency_1s', true);
const rollingP95Latency1s = new Trend('rolling_p95_latency_1s', true);

// Rolling window metrics - TTFB (0.1s, 0.5s, 1s)
const rollingAvgTTFB01s = new Trend('rolling_avg_ttfb_0_1s', true);
const rollingP95TTFB01s = new Trend('rolling_p95_ttfb_0_1s', true);
const rollingAvgTTFB05s = new Trend('rolling_avg_ttfb_0_5s', true);
const rollingP95TTFB05s = new Trend('rolling_p95_ttfb_0_5s', true);
const rollingAvgTTFB1s = new Trend('rolling_avg_ttfb_1s', true);
const rollingP95TTFB1s = new Trend('rolling_p95_ttfb_1s', true);

// Rolling window metrics - Cache Hit Ratio (0.1s, 0.5s, 1s)
const rollingHitRatio01s = new Trend('rolling_hit_ratio_0_1s', true);
const rollingHitRatio05s = new Trend('rolling_hit_ratio_0_5s', true);
const rollingHitRatio1s = new Trend('rolling_hit_ratio_1s', true);

// Rolling request counter per second
const rollingReqCount1s = new Counter('rolling_req_count_1s');

// Corpus of endpoints available on the mock origin
export const PATHS = [
    'binary_100kb',
    'large_css',
    'large_html',
    'large_image',
    'large_js',
    'large_json',
    'medium_image',
    'medium_json',
    'small_css',
    'small_html',
    'small_image',
    'small_js',
    'small_json',
];

// Alias for compatibility
export const FILES = PATHS;

// Target cache key cardinality (can be tuned here or via CARDINALITY env var)
export const CARDINALITY_TARGET = Number(__ENV.CARDINALITY) || 1000;

// Target container host and port (configurable via TARGET_URL, BASE_URL, TARGET_HOST, HOST, PORT)
const TARGET_URL = __ENV.TARGET_URL || __ENV.BASE_URL;
const TARGET_HOST = __ENV.TARGET_HOST || __ENV.HOST || 'caddy-mak';
const PORT = __ENV.PORT || '80';
const BASE_URL = TARGET_URL || `http://${TARGET_HOST}:${PORT}`;

// Helper to parse k6 duration strings (e.g. '5s', '1m', '2m30s') into total seconds
function parseDurationSec(durationStr: string): number {
    let totalSec = 0;
    const matches = durationStr.match(/(\d+)(h|m|s)/g);
    if (!matches) {
        const num = parseInt(durationStr, 10);
        return isNaN(num) ? 0 : num;
    }
    for (const match of matches) {
        const unit = match.slice(-1);
        const val = parseInt(match.slice(0, -1), 10);
        if (unit === 's') totalSec += val;
        else if (unit === 'm') totalSec += val * 60;
        else if (unit === 'h') totalSec += val * 3600;
    }
    return totalSec;
}

export const STAGES = [
    {duration: '1s', target: 10},   // Warm-up ramp
    {duration: '10s', target: 50},   // Medium load
    {duration: '5s', target: 200},  // High load spike
    {duration: '5s', target: 50},   // Step down
    {duration: '2s', target: 100},  // Quickly ramp up
    {duration: '2s', target: 350},  // Continue
    {duration: '2s', target: 450},  // Super peak
    {duration: '2s', target: 50},   // Step down
    {duration: '2s', target: 0},    // Cool down
];

export const TOTAL_DURATION_SEC = STAGES.reduce((acc, stage) => acc + parseDurationSec(stage.duration), 0);

// Build thresholds map to preserve submetric time-series tags dynamically
const thresholdsConfig: Record<string, string[]> = {
    http_req_failed: ['rate<0.01'],
};

for (let s = 0; s < TOTAL_DURATION_SEC; s++) {
    thresholdsConfig[`rolling_avg_latency_1s{sec:${s}}`] = [];
    thresholdsConfig[`rolling_p95_latency_1s{sec:${s}}`] = [];
    thresholdsConfig[`rolling_avg_ttfb_1s{sec:${s}}`] = [];
    thresholdsConfig[`rolling_p95_ttfb_1s{sec:${s}}`] = [];
    thresholdsConfig[`rolling_hit_ratio_1s{sec:${s}}`] = [];
    thresholdsConfig[`rolling_req_count_1s{sec:${s}}`] = [];
}

export const options: Options = {
    scenarios: {
        benchmark_ramp: {
            executor: 'ramping-vus',
            startVUs: 1,
            stages: STAGES,
            gracefulRampDown: '0s',
        },
    },
    thresholds: thresholdsConfig,
};

interface MetricEntry {
    time: number;
    duration: number; // Latency in ms
    waiting: number;  // TTFB in ms
    isHit: boolean;   // Cache hit status
}

// Sliding window history maintained per VU instance
const history: MetricEntry[] = [];

function calcStats(entries: MetricEntry[], getVal: (e: MetricEntry) => number) {
    if (entries.length === 0) return {avg: 0, p95: 0, hitRatio: 0};
    const vals = entries.map(getVal).sort((a, b) => a - b);
    const sum = vals.reduce((acc, v) => acc + v, 0);
    const avg = sum / vals.length;
    const p95Idx = Math.max(0, Math.ceil(0.95 * vals.length) - 1);
    const p95 = vals[p95Idx];
    const hitCount = entries.filter((e) => e.isHit).length;
    const hitRatio = (hitCount / entries.length) * 100;
    return {avg, p95, hitRatio};
}

function checkCacheHit(res: any): boolean {
    if (!res || !res.headers) return false;

    // Check Cache-Status header (Caddy, Souin, mak-cache)
    const cacheStatusKey = Object.keys(res.headers).find((k) => k.toLowerCase() === 'cache-status');
    if (cacheStatusKey && String(res.headers[cacheStatusKey]).toLowerCase().includes('hit')) {
        return true;
    }

    // Check X-Cache header (Varnish, Squid)
    const xCacheKey = Object.keys(res.headers).find((k) => k.toLowerCase() === 'x-cache');
    if (xCacheKey && String(res.headers[xCacheKey]).toLowerCase().includes('hit')) {
        return true;
    }

    // Check X-Cache-Status header (Nginx)
    const xCacheStatusKey = Object.keys(res.headers).find((k) => k.toLowerCase() === 'x-cache-status');
    if (xCacheStatusKey && String(res.headers[xCacheStatusKey]).toLowerCase().includes('hit')) {
        return true;
    }

    // Check X-Cache-Lookup header (Squid)
    const xCacheLookupKey = Object.keys(res.headers).find((k) => k.toLowerCase() === 'x-cache-lookup');
    if (xCacheLookupKey && String(res.headers[xCacheLookupKey]).toLowerCase().includes('hit')) {
        return true;
    }

    // Check Age header > 0
    const ageKey = Object.keys(res.headers).find((k) => k.toLowerCase() === 'age');
    if (ageKey) {
        const ageVal = parseInt(String(res.headers[ageKey]), 10);
        if (!isNaN(ageVal) && ageVal > 0) {
            return true;
        }
    }

    return false;
}

export default function () {
    const now = Date.now();
    const elapsedSec = Math.min(
        TOTAL_DURATION_SEC - 1,
        Math.max(0, Math.floor(exec.instance.currentTestRunDuration / 1000))
    );

    // Procedural and deterministic endpoint selection
    const iteration = exec.scenario.iterationInTest;
    const path = PATHS[iteration % PATHS.length];

    // Procedural and deterministic cache key cardinality generation
    const keyId = (iteration * 2654435761) % CARDINALITY_TARGET;
    const url = `${BASE_URL}/${path}?k=${keyId}`;

    const res = http.get(url, {
        tags: { name: `/${path}` },
        responseCallback: http.expectedStatuses(200, 404),
    });

    check(res, {
        'status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    });

    const duration = res.timings.duration;
    const waiting = res.timings.waiting; // TTFB
    const isHit = checkCacheHit(res);

    history.push({time: now, duration, waiting, isHit});

    // Evict items older than 1 second (1000ms)
    while (history.length > 0 && history[0].time < now - 1000) {
        history.shift();
    }

    // Tag for time-series extraction per elapsed second
    const tags = {sec: `${elapsedSec}`};

    // 0.1s (100ms) window
    const window01s = history.filter((e) => e.time >= now - 100);
    const statsLat01s = calcStats(window01s, (e) => e.duration);
    const statsTTFB01s = calcStats(window01s, (e) => e.waiting);
    rollingAvgLatency01s.add(statsLat01s.avg, tags);
    rollingP95Latency01s.add(statsLat01s.p95, tags);
    rollingAvgTTFB01s.add(statsTTFB01s.avg, tags);
    rollingP95TTFB01s.add(statsTTFB01s.p95, tags);
    rollingHitRatio01s.add(statsLat01s.hitRatio, tags);

    // 0.5s (500ms) window
    const window05s = history.filter((e) => e.time >= now - 500);
    const statsLat05s = calcStats(window05s, (e) => e.duration);
    const statsTTFB05s = calcStats(window05s, (e) => e.waiting);
    rollingAvgLatency05s.add(statsLat05s.avg, tags);
    rollingP95Latency05s.add(statsLat05s.p95, tags);
    rollingAvgTTFB05s.add(statsTTFB05s.avg, tags);
    rollingP95TTFB05s.add(statsTTFB05s.p95, tags);
    rollingHitRatio05s.add(statsLat05s.hitRatio, tags);

    // 1.0s (1000ms) window
    const window1s = history;
    const statsLat1s = calcStats(window1s, (e) => e.duration);
    const statsTTFB1s = calcStats(window1s, (e) => e.waiting);
    rollingAvgLatency1s.add(statsLat1s.avg, tags);
    rollingP95Latency1s.add(statsLat1s.p95, tags);
    rollingAvgTTFB1s.add(statsTTFB1s.avg, tags);
    rollingP95TTFB1s.add(statsTTFB1s.p95, tags);
    rollingHitRatio1s.add(statsLat1s.hitRatio, tags);
    rollingReqCount1s.add(1, tags);
}

const SPARK_CHARS = [' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'];

function generateSparkline(values: number[], minVal?: number, maxVal?: number): string {
    if (values.length === 0) return '';
    const min = minVal !== undefined ? minVal : Math.min(...values);
    const max = maxVal !== undefined ? maxVal : Math.max(...values);
    const range = max - min;

    return values
        .map((v) => {
            if (range === 0) return SPARK_CHARS[0];
            const idx = Math.min(
                SPARK_CHARS.length - 1,
                Math.max(0, Math.floor(((v - min) / range) * (SPARK_CHARS.length - 1)))
            );
            return SPARK_CHARS[idx];
        })
        .join('');
}

export function handleSummary(data: any) {
    const csvRows = ['sec,reqs_per_sec,avg_latency_ms,p95_latency_ms,avg_ttfb_ms,p95_ttfb_ms,hit_ratio_pct'];
    const timeSeriesData: Array<{
        sec: number;
        reqs: number;
        avgLat: number;
        p95Lat: number;
        avgTTFB: number;
        p95TTFB: number;
        hitRatio: number;
    }> = [];

    const secKeys = Object.keys(data.metrics)
        .map((k) => {
            const match = k.match(/sec:(\d+)/);
            return match ? parseInt(match[1], 10) : -1;
        })
        .filter((s) => s >= 0);

    const totalSecs = secKeys.length > 0 ? Math.max(...secKeys) + 1 : TOTAL_DURATION_SEC;

    for (let s = 0; s < totalSecs; s++) {
        const keyReqs = `rolling_req_count_1s{sec:${s}}`;
        const keyLatAvg = `rolling_avg_latency_1s{sec:${s}}`;
        const keyLatP95 = `rolling_p95_latency_1s{sec:${s}}`;
        const keyTTFBAvg = `rolling_avg_ttfb_1s{sec:${s}}`;
        const keyTTFBP95 = `rolling_p95_ttfb_1s{sec:${s}}`;
        const keyHitRatio = `rolling_hit_ratio_1s{sec:${s}}`;

        const reqs = data.metrics[keyReqs]?.values?.count || 0;
        const avgLat = data.metrics[keyLatAvg]?.values?.avg || 0;
        const p95Lat = data.metrics[keyLatP95]?.values?.avg || 0;
        const avgTTFB = data.metrics[keyTTFBAvg]?.values?.avg || 0;
        const p95TTFB = data.metrics[keyTTFBP95]?.values?.avg || 0;
        const hitRatio = data.metrics[keyHitRatio]?.values?.avg || 0;

        csvRows.push(`${s},${reqs},${avgLat.toFixed(2)},${p95Lat.toFixed(2)},${avgTTFB.toFixed(2)},${p95TTFB.toFixed(2)},${hitRatio.toFixed(1)}`);
        timeSeriesData.push({sec: s, reqs, avgLat, p95Lat, avgTTFB, p95TTFB, hitRatio});
    }

    const csvOutput = csvRows.join('\n') + '\n';

    const reqsList = timeSeriesData.map((d) => d.reqs);
    const steadyReqsList = reqsList.length > 1 ? reqsList.slice(1) : reqsList;
    const avgLats = timeSeriesData.map((d) => d.avgLat);
    const p95Lats = timeSeriesData.map((d) => d.p95Lat);
    const avgTTFBs = timeSeriesData.map((d) => d.avgTTFB);
    const p95TTFBs = timeSeriesData.map((d) => d.p95TTFB);
    const hitRatios = timeSeriesData.map((d) => d.hitRatio);

    const sparkReqs = generateSparkline(steadyReqsList);
    const sparkAvgLat = generateSparkline(avgLats);
    const sparkP95Lat = generateSparkline(p95Lats);
    const sparkAvgTTFB = generateSparkline(avgTTFBs);
    const sparkP95TTFB = generateSparkline(p95TTFBs);
    const sparkHitRatio = generateSparkline(hitRatios, 0, 100);

    const formatRange = (arr: number[], unit: string) => {
        if (arr.length === 0) return `[Min: 0.00${unit}, Max: 0.00${unit}]`;
        const min = Math.min(...arr).toFixed(2);
        const max = Math.max(...arr).toFixed(2);
        return `[Min: ${min}${unit}, Max: ${max}${unit}]`;
    };

    const formatPctRange = (arr: number[]) => {
        if (arr.length === 0) return '[Min: 0.0%, Max: 0.0%]';
        const min = Math.min(...arr).toFixed(1);
        const max = Math.max(...arr).toFixed(1);
        return `[Min: ${min}%, Max: ${max}%]`;
    };

    const totalReqs = data.metrics.http_reqs?.values?.count || 0;
    const avgRps = data.metrics.http_reqs?.values?.rate || 0;
    const overallAvgLat = data.metrics.http_req_duration?.values?.avg || 0;
    const overallP95Lat = data.metrics.http_req_duration?.values?.['p(95)'] || 0;
    const peakP95Lat = p95Lats.length > 0 ? Math.max(...p95Lats) : 0;
    const overallAvgTTFB = data.metrics.http_req_waiting?.values?.avg || 0;
    const overallP95TTFB = data.metrics.http_req_waiting?.values?.['p(95)'] || 0;
    const peakP95TTFB = p95TTFBs.length > 0 ? Math.max(...p95TTFBs) : 0;
    const avgHitRatio = hitRatios.length > 0 ? hitRatios.reduce((a, b) => a + b, 0) / hitRatios.length : 0;
    const maxHitRatio = hitRatios.length > 0 ? Math.max(...hitRatios) : 0;

    let peakRps = 0;
    let peakSec = 0;
    for (let i = 0; i < timeSeriesData.length; i++) {
        if (timeSeriesData[i].sec < 3 && timeSeriesData.length > 3) continue; // skip initial VU spin-up burst
        if (timeSeriesData[i].reqs > peakRps) {
            peakRps = timeSeriesData[i].reqs;
            peakSec = timeSeriesData[i].sec;
        }
    }

    let mockOriginCountStr = 'N/A';
    try {
        const mockUrl = __ENV.MOCK_URL || __ENV.MOCK_ORIGIN || (BASE_URL.includes('mockorigin') ? BASE_URL : 'http://mockorigin:80');
        let res = http.get(`${mockUrl}/_requests`);
        if (res.status === 200) {
            const body = res.json() as any;
            if (body && typeof body.count === 'number') {
                mockOriginCountStr = `${body.count.toLocaleString()} reqs`;
            } else if (!isNaN(parseInt(res.body as string, 10))) {
                mockOriginCountStr = `${parseInt(res.body as string, 10).toLocaleString()} reqs`;
            }

            // Clear count from mock server after reading
            http.del(`${mockUrl}/_requests`);
        }
    } catch (_) {
        // Ignore errors when querying mock server count
    }

    const reportBox = [
        '  Charts',
        '  --------------------------------------------------------------------------------------------------',
        `  Avg Latency (ms):   ${sparkAvgLat}  ${formatRange(avgLats, 'ms')}`,
        `  P95 Latency (ms):   ${sparkP95Lat}  ${formatRange(p95Lats, 'ms')}`,
        `  Avg TTFB (ms):      ${sparkAvgTTFB}  ${formatRange(avgTTFBs, 'ms')}`,
        `  P95 TTFB (ms):      ${sparkP95TTFB}  ${formatRange(p95TTFBs, 'ms')}`,
        `  Hit Ratio (%):      ${sparkHitRatio}  ${formatPctRange(hitRatios)}`,
        '',
        '  Summary',
        '  --------------------------------------------------------------------------------------------------',
        `  • Total Completed Requests:  ${totalReqs.toLocaleString()} reqs`,
        `  • Total Origin Requests:     ${mockOriginCountStr}`,
        `  • Throughput (req/s):        Overall Avg: ${avgRps.toFixed(1)} req/s  |  Peak 1s Window: ${peakRps.toLocaleString()} req/s (at sec ${peakSec})`,
        `  • Latency (Duration):        Overall Avg: ${overallAvgLat.toFixed(2)} ms  |  Overall P95: ${overallP95Lat.toFixed(2)} ms  (Peak 1s P95: ${peakP95Lat.toFixed(2)} ms)`,
        `  • Time To First Byte (TTFB): Overall Avg: ${overallAvgTTFB.toFixed(2)} ms  |  Overall P95: ${overallP95TTFB.toFixed(2)} ms  (Peak 1s P95: ${peakP95TTFB.toFixed(2)} ms)`,
        `  • Cache Hit Ratio:           Overall Avg: ${avgHitRatio.toFixed(1)}%  |  Peak: ${maxHitRatio.toFixed(1)}%`,
    ];

    console.log(reportBox.join('\n'));

    return {
        '/benchmarks/benchmark_results.csv': csvOutput,
    };
}

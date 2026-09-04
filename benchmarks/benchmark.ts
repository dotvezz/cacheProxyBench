import exec from 'k6/execution';
import { Options } from 'k6/options';
import {
    PATHS,
    CARDINALITY_TARGET,
    BASE_URL,
    parseDurationSec,
    createThresholds,
    executeRequest,
    handleSummary as commonHandleSummary,
    ZipfianGenerator,
    MetricEntry,
} from './common.ts';

// Zipf skew parameter (s or alpha). Real-world web/CDN caching traces typically range from 0.7 to 1.0.
// Default is 0.85 (e.g. top ~10-15% of items generate ~70% of traffic).
export const ZIPF_EXPONENT = Number(__ENV.ZIPF_EXPONENT) || 0.85;

// Whether to generate requests deterministically based on test iteration index
export const IS_DETERMINISTIC = __ENV.DETERMINISTIC === 'true' || __ENV.DETERMINISTIC === '1';

// Precomputed Zipfian samplers
const keyZipf = new ZipfianGenerator(CARDINALITY_TARGET, ZIPF_EXPONENT);
const pathZipf = new ZipfianGenerator(PATHS.length, 0.75);

// Realistic benchmark staging:
// 1. Warm-up (cache filling phase)
// 2. Ramp to sustained operational load
// 3. Sustained steady-state traffic
// 4. Peak burst traffic
// 5. Recovery & Cool-down
export const STAGES = [
    { duration: '5s', target: 25 },   // Cache warm-up
    { duration: '5s', target: 100 },  // Ramp to operational load
    { duration: '5s', target: 200 },  // Sustained steady state
    { duration: '2s', target: 400 }, // Peak burst load
    { duration: '2s', target: 1000 }, // Peak burst load
    { duration: '5s', target: 50 },   // Recovery
    { duration: '5s', target: 0 },    // Cool down
];

export const TOTAL_DURATION_SEC = STAGES.reduce((acc, stage) => acc + parseDurationSec(stage.duration), 0);

export const options: Options = {
    scenarios: {
        zipfian_benchmark: {
            executor: 'ramping-vus',
            startVUs: 1,
            stages: STAGES,
            gracefulRampDown: '0s',
        },
    },
    thresholds: createThresholds(TOTAL_DURATION_SEC),
};

// Sliding window history maintained per VU instance
const history: MetricEntry[] = [];

export default function () {
    const iteration = exec.scenario.iterationInTest;

    // Sample endpoint path and cache key following Zipfian power-law distribution
    let pathIndex: number;
    let keyId: number;

    if (IS_DETERMINISTIC) {
        pathIndex = pathZipf.nextFromIteration(iteration);
        keyId = keyZipf.nextFromIteration(iteration + 99991);
    } else {
        pathIndex = pathZipf.next();
        keyId = keyZipf.next();
    }

    const path = PATHS[pathIndex];
    const url = `${BASE_URL}/${path}?k=${keyId}`;

    executeRequest(url, path, TOTAL_DURATION_SEC, history);
}

export function handleSummary(data: any) {
    return commonHandleSummary(data, TOTAL_DURATION_SEC, '/benchmarks/benchmark_results.csv');
}

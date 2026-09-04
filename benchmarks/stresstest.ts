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
    MetricEntry,
} from './common.ts';

// Stress test staging: rapid, aggressive spikes up to 2,000 VUs
export const STAGES = [
    { duration: '1s', target: 10 },   // Warm-up ramp
    { duration: '2s', target: 50 },   // Medium load
    { duration: '2s', target: 200 },  // High load spike
    { duration: '2s', target: 50 },   // Step down
    { duration: '2s', target: 100 },  // Quickly ramp up
    { duration: '2s', target: 450 },  // Continue
    { duration: '2s', target: 2000 }, // Super peak
    { duration: '2s', target: 50 },   // Step down
    { duration: '2s', target: 0 },    // Cool down
];

export const TOTAL_DURATION_SEC = STAGES.reduce((acc, stage) => acc + parseDurationSec(stage.duration), 0);

export const options: Options = {
    scenarios: {
        stresstest_ramp: {
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
    const path = PATHS[iteration % PATHS.length];

    // Procedural and deterministic cache key cardinality generation
    const keyId = (iteration * 2654435761) % CARDINALITY_TARGET;
    const url = `${BASE_URL}/${path}?k=${keyId}`;

    executeRequest(url, path, TOTAL_DURATION_SEC, history);
}

export function handleSummary(data: any) {
    return commonHandleSummary(data, TOTAL_DURATION_SEC, '/benchmarks/stresstest_results.csv');
}

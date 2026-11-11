import http from "k6/http";
import {check} from "k6";

export const options = {
    stages: [
        { duration: '10s', target: 20},
        { duration: '30s', target: 20},
        { duration: '10s', target: 100},
        { duration: '10s', target: 0}
    ],
    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(95)<200'],
    }
}

export default function() {
    const url = 'http://localhost:4000/jobs';

    const res = http.post(
        url,
        JSON.stringify({ type: 'echo', payload: { msg: 'hi' } }),
        { headers: { 'Content-Type': 'application/json' } }
    );

    check(res, 
        { 'accepted: (202)': (r) => r.status === 202}
    )
}
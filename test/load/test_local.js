import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10,
    duration: '30s'
};

export default function () {
    const res = http.get('http://localhost:8080/test1_testhttp');

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 100ms': (r) => r.timings.duration < 100,
    });

    sleep(1)
}
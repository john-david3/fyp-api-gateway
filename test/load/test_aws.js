import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
	vus: 1000,
	duration: '30s',
};

export default function (){
	const res = http.get('https://synchro-accelerator.com/takehiro_load');

	check(res, {
		'status is 200': (r) => r.status === 200,
		'response time < 100ms': (r) => r.timings.duration < 100,
	});
}

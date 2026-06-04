import http from 'k6/http';
import { sleep, check } from 'k6';
import exec from 'k6/execution';

// Move to docker 
export const options = {
  scenarios: {
    default: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        // 15 minute run range
        { duration: '2m', target: 20 },
        { duration: '4m', target: 50 },
        { duration: '1m', target: 1 },
        { duration: '5m', target: 30 },
        { duration: '3m', target: 5 },
      ],
    },
  },
};

// Seed 100 links that can be used by the main VUs
export function setup() {
  const paths = [];
  for (let i = 0; i < 100; i++) {
    const res = http.post(`http://${__ENV.TARGET_URL}/api/link`,
      JSON.stringify({ origin_url: `https://seed-${i}.aaroncunliffe.dev` }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    if (res.status === 201) {
      paths.push(res.json().data.short_path);
    }

    sleep(0.10)
  }
  return { paths, firstPath: paths[0] };
}

// Very basic synthetic load tests to drive dashboard statistics
export default function (data) {
  let roll = Math.random();

  if (roll < 0.10) {
    let res = http.post(`http://${__ENV.TARGET_URL}/api/link`,
      JSON.stringify({ origin_url: `https://site-${exec.vu.idInTest}.aaroncunliffe.dev` }),
      { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'created': (r) => r.status === 201 });
  }
  else if (roll < 0.90) {
    // Redirect to known path
    const path = data.paths[Math.floor(Math.random() * data.paths.length)];
    const res = http.get(`http://${__ENV.TARGET_URL}/${path}`, { redirects: 0 });
    check(res, { 'redirect': (r) => r.status === 302 });
  }
  else if (roll < 0.95) {
    // Redirect to an unknown path
    const res = http.get(`http://${__ENV.TARGET_URL}/not-found`, { redirects: 0 });
    check(res, { 'not-found': (r) => r.status === 404 });
  }
  else {
    // Use an already in use path
    const path = data.paths[Math.floor(Math.random() * data.paths.length)];
    let res = http.post(`http://${__ENV.TARGET_URL}/api/link`,
      JSON.stringify({ origin_url: `https://site-${exec.vu.idInTest}.aaroncunliffe.dev`, short_path: `${path}` }),
      { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'conflict': (r) => r.status === 409 });
  }


  sleep(1);
}
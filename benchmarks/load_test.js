import http from 'k6/http';
import { check } from 'k6';

export const options = {
    // Define the number of iterations for the test
    // iterations: 10, // number of iterations for default function per test (either iterations or stages!)
    // duration: '1m', // duration of test (either duration or stages!)
    vus: 10000, // amount of virtual users
    stages: [ // variable amount of virtual users by stages
        { duration: '15s', target: 2000 },
        { duration: '45s', target: 10000 },
        { duration: '55s', target: 4000 },
    ],
};

export default function () {
    const serverAddr = 'http://localhost:8080';

    const shortIDs = [];
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    for (let i = 0; i < 10; i++) {
        // Генерируем уникальный URL, добавляя timestamp к базовому URL
        const uniqueUrl = `https://example.com/some/path?id=${Date.now()}-${__VU}-${__ITER}-${i}`;
        const payload = JSON.stringify({
            url: uniqueUrl,
        });

        const res = http.post(`${serverAddr}/api/shorten`, payload, params);

        // Проверяем, что запрос прошел успешно (200 OK или 201 Created)
        check(res, {
            'status is 201 or 200': (r) => r.status === 201 || r.status === 200,
        });

        const shortID = res.json().result;
        shortIDs.push(shortID);
    }

    http.get(`${serverAddr}/api/user/urls`);

    const payload = JSON.stringify(shortIDs.slice(0, 5));

    http.del(`${serverAddr}/api/user/urls`, payload, params)
}

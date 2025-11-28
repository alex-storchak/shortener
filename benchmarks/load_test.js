import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    duration: '2m',
    vus: 100, // Уменьшенное количество VUs
    // thresholds: {
    //     http_req_failed: ['rate<0.01'], // Меньше 1% ошибок
    // },
};

export default function () {
    const serverAddr = 'http://localhost:8080';
    const params = {
        headers: { 'Content-Type': 'application/json' },
    };

    // Только одно создание URL за итерацию
    const uniqueUrl = `https://test.com/url/${__VU}-${__ITER}-${Date.now()}`;
    const createRes = http.post(`${serverAddr}/api/shorten`,
        JSON.stringify({ url: uniqueUrl }), params);

    // check(createRes, {
    //     'create success': (r) => r.status === 200 || r.status === 201,
    // });

    // Периодически получаем список URL (каждая 5-я итерация)
    if (__ITER % 5 === 0) {
        http.get(`${serverAddr}/api/user/urls`);
    }

    // Периодически удаляем (каждая 10-я итерация)
    if (__ITER % 10 === 0 && createRes.status === 201) {
        const shortID = createRes.json().result;
        http.del(`${serverAddr}/api/user/urls`,
            JSON.stringify([shortID]), params);
    }

    // sleep(1); // Пауза между итерациями
}
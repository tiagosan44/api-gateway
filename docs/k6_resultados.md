Interpretación de resultados de k6

Este documento explica e interpreta los resultados mostrados por k6 en el reporte compartido al ejecutar `k6 run tests/load/k6_test.js` con 3 etapas (10 → 50 VUs → 0) durante ~2 minutos.

Resumen ejecutivo
- Estado: PRUEBA EXITOSA. Todos los checks pasaron (100% éxito) y no hubo errores HTTP.
- Latencia: p95 = 4.22 ms, muy por debajo del umbral definido (p95 < 500 ms).
- Error rate: 0.00% (umbral < 10%).
- Tráfico: ~44.66 req/s en promedio (5,388 requests totales), limitado principalmente por el `sleep(1)` del script.
- Concurrencia: se alcanzaron 50 VUs, con 2,694 iteraciones (cada iteración hace 2 requests: /health y /metrics).

Conclusión: El gateway responde de forma consistente y extremadamente rápida en las rutas `/health` y `/metrics` bajo la carga configurada. No hay señales de saturación en este escenario.

Lectura del reporte
1) THRESHOLDS (umbrales)
- errors: ✓ 'rate<0.1' → rate=0.00%
- http_req_duration: ✓ 'p(95)<500' → p95=4.22 ms
Interpretación: Ambas condiciones definidas en `tests/load/k6_test.js` se cumplen holgadamente. El 95% de las peticiones terminó en menos de 4.22 ms (vs. el límite de 500 ms) y la tasa de errores es 0% (vs. el límite de 10%).

2) TOTAL RESULTS (checks)
- checks_total: 8,082
- checks_succeeded: 100.00% (8,082/8,082)
Checks en el script: “health status is 200”, “health response time < 100ms”, “metrics status is 200” → todos OK.

3) HTTP
- http_req_duration: avg=2.07ms, p90=3.36ms, p95=4.22ms, max=11.93ms
- http_req_failed: 0.00% (0 de 5,388)
- http_reqs: 5,388 (≈ 44.66 req/s)
Interpretación: Latencia muy baja y estable, sin errores.

4) EXECUTION
- iterations: 2,694 (≈ 22.33 iter/s)
- iteration_duration: ≈ 1s constante (p95=1s)
- VUs: min=1, max=50
Cada iteración hace 2 requests y `sleep(1)`. Por eso el RPS está limitado por el `sleep(1)`, no por el gateway.

5) NETWORK
- data_received: 19 MB totales (~157 kB/s)
- data_sent: 412 kB totales (~3.4 kB/s)
Valores esperables para endpoints ligeros.

¿Qué nos dicen estos números del sistema?
- Capacidad: Maneja sin esfuerzo ~45 req/s en estas rutas con p95 ≈ 4 ms; la limitación es el script.
- Estabilidad: 0% de errores; latencia plana.
- Overhead bajo: /health y /metrics son baratos; no refleja comportamiento de proxy a downstreams.

Recomendaciones y siguientes pasos
1) Reducir o quitar `sleep(1)` para incrementar el RPS (p. ej., `sleep(0.1)`).
2) Aumentar VUs y duración (p. ej., hasta 200 VUs) para observar picos y estabilidad.
3) Probar rutas de negocio (`/v1/{service}/{path}`) con upstream real/simulado para medir E2E.
4) Añadir umbrales adicionales: p99, `http_req_failed<1%`, checks de contenido.
5) Separar escenarios por endpoint para aislar efectos.
6) Considerar autenticación (JWT/OIDC) con tokens válidos si aplican.

Cómo replicar la prueba actual
- docker-compose up -d; k6 run tests/load/k6_test.js
- O: BASE_URL=http://localhost:8080 k6 run tests/load/k6_test.js
- Perfil de test: docker-compose -f docker-compose.test.yml up -d; BASE_URL=http://localhost:8082 k6 run tests/load/k6_test.js

Glosario rápido
- VU: usuario virtual.
- Iteration: ejecución de la función `default`.
- RPS: requests por segundo.
- p95: 95% de las peticiones son más rápidas que ese valor.
- Threshold: umbral que debe cumplirse (p. ej., p95 < 500 ms).
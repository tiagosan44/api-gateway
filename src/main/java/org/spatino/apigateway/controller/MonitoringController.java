package org.spatino.apigateway.controller;

import org.spatino.apigateway.generated.api.MonitoringApi;
import org.spatino.apigateway.generated.model.HealthResponse;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.Map;

@RestController
public class MonitoringController implements MonitoringApi {

    @Override
    public Mono<ResponseEntity<HealthResponse>> healthCheck(ServerWebExchange exchange) {
        HealthResponse health = new HealthResponse();
        health.setStatus(HealthResponse.StatusEnum.UP);
        health.setTimestamp(OffsetDateTime.now());

        Map<String, String> upstreamServices = new HashMap<>();
        upstreamServices.put("product-service", "UP");
        upstreamServices.put("inventory-service", "UP");
        health.setUpstreamServices(upstreamServices);

        return Mono.just(ResponseEntity.ok(health));
    }
}

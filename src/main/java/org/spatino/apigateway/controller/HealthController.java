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

/**
 * Health check controller implementing the generated MonitoringApi interface.
 */
@RestController
public class HealthController implements MonitoringApi {

    @Override
    public Mono<ResponseEntity<HealthResponse>> healthCheck(ServerWebExchange exchange) {
        HealthResponse response = new HealthResponse();
        response.setStatus(HealthResponse.StatusEnum.UP);
        response.setTimestamp(OffsetDateTime.now());

        // TODO: Check actual upstream services
        Map<String, String> upstreamServices = new HashMap<>();
        upstreamServices.put("product-service", "UP");
        upstreamServices.put("inventory-service", "UP");
        response.setUpstreamServices(upstreamServices);

        return Mono.just(ResponseEntity.ok(response));
    }
}

package org.spatino.apigateway.controller;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.api.InventoryApi;
import org.spatino.apigateway.generated.model.Inventory;
import org.spatino.apigateway.generated.model.InventoryUpdate;
import org.spatino.apigateway.service.InventoryService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

import java.util.UUID;

@Slf4j
@RestController
@RequiredArgsConstructor
public class InventoryController implements InventoryApi {

    private final InventoryService inventoryService;

    @Override
    public Mono<ResponseEntity<Inventory>> getInventory(UUID productId, ServerWebExchange exchange) {
        return inventoryService.getInventory(productId)
                .map(inventory -> ResponseEntity.ok()
                        .header("X-Cache-Status", "MISS")
                        .body(inventory));
    }

    @Override
    public Mono<ResponseEntity<Inventory>> updateInventory(UUID productId, Mono<InventoryUpdate> inventoryUpdate, ServerWebExchange exchange) {
        return inventoryUpdate.flatMap(update ->
                inventoryService.updateInventory(productId, update)
                    .map(ResponseEntity::ok)
        );
    }
}

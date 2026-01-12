package org.spatino.apigateway.service;

import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.model.Inventory;
import org.spatino.apigateway.generated.model.InventoryUpdate;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.time.OffsetDateTime;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
public class InventoryService {

    private final Map<UUID, Inventory> inventoryStore = new ConcurrentHashMap<>();

    public Mono<Inventory> getInventory(UUID productId) {
        log.info("Getting inventory for product: {}", productId);
        return Mono.fromCallable(() -> {
            Inventory inventory = inventoryStore.get(productId);
            if (inventory == null) {
                // Create default inventory if not exists
                inventory = new Inventory();
                inventory.setProductId(productId);
                inventory.setQuantity(100); // Default stock
                inventory.setReservedQuantity(0);
                inventory.setAvailable(true);
                inventory.setLastUpdated(OffsetDateTime.now());
                inventoryStore.put(productId, inventory);
            }
            return inventory;
        });
    }

    public Mono<Inventory> updateInventory(UUID productId, InventoryUpdate inventoryUpdate) {
        log.info("Updating inventory for product: {}", productId);
        return Mono.fromCallable(() -> {
            Inventory inventory = inventoryStore.get(productId);
            if (inventory == null) {
                inventory = new Inventory();
                inventory.setProductId(productId);
                inventory.setQuantity(0);
                inventory.setReservedQuantity(0);
                inventory.setAvailable(false);
            }

            if (inventoryUpdate.getQuantity() != null) {
                inventory.setQuantity(inventoryUpdate.getQuantity());
            }
            if (inventoryUpdate.getReservedQuantity() != null) {
                inventory.setReservedQuantity(inventoryUpdate.getReservedQuantity());
            }

            // Update availability based on quantity
            int availableQuantity = inventory.getQuantity() - (inventory.getReservedQuantity() != null ? inventory.getReservedQuantity() : 0);
            inventory.setAvailable(availableQuantity > 0);
            inventory.setLastUpdated(OffsetDateTime.now());

            inventoryStore.put(productId, inventory);
            return inventory;
        });
    }
}

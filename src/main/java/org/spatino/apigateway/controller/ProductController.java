package org.spatino.apigateway.controller;

import lombok.RequiredArgsConstructor;
import org.spatino.apigateway.generated.api.ProductsApi;
import org.spatino.apigateway.generated.model.Product;
import org.spatino.apigateway.generated.model.ProductListResponse;
import org.spatino.apigateway.generated.model.ProductUpdate;
import org.spatino.apigateway.service.ProductService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

import java.math.BigDecimal;
import java.util.UUID;

@RestController
@RequiredArgsConstructor
public class ProductController implements ProductsApi {

    private final ProductService productService;

    @Override
    public Mono<ResponseEntity<ProductListResponse>> listProducts(
            Integer page, Integer size, String category,
            BigDecimal minPrice, BigDecimal maxPrice, String search,
            ServerWebExchange exchange) {
        
        return productService.listProducts(page, size, category, minPrice, maxPrice, search)
                .map(response -> ResponseEntity.ok()
                        .header("X-RateLimit-Limit", "100")
                        .header("X-RateLimit-Remaining", "95")
                        .header("X-RateLimit-Reset", String.valueOf(System.currentTimeMillis() / 1000 + 3600))
                        .header("X-Cache-Status", "MISS")
                        .body(response));
    }

    @Override
    public Mono<ResponseEntity<Product>> getProduct(UUID productId, ServerWebExchange exchange) {
        return productService.getProduct(productId)
                .map(product -> ResponseEntity.ok()
                        .header("X-Cache-Status", "MISS")
                        .body(product))
                .defaultIfEmpty(ResponseEntity.notFound().build());
    }

    @Override
    public Mono<ResponseEntity<Product>> updateProduct(UUID productId, Mono<ProductUpdate> productUpdate, ServerWebExchange exchange) {
        return productUpdate.flatMap(update ->
                productService.updateProduct(productId, update)
                    .map(ResponseEntity::ok)
                    .defaultIfEmpty(ResponseEntity.notFound().build())
        );
    }

    @Override
    public Mono<ResponseEntity<Void>> deleteProduct(UUID productId, ServerWebExchange exchange) {
        return productService.deleteProduct(productId)
                .map(deleted -> deleted ? 
                        ResponseEntity.noContent().<Void>build() : 
                        ResponseEntity.notFound().<Void>build());
    }
}

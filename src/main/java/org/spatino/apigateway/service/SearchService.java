package org.spatino.apigateway.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.model.ProductListResponse;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

@Slf4j
@Service
@RequiredArgsConstructor
public class SearchService {

    private final ProductService productService;

    public Mono<ProductListResponse> searchProducts(String query, Integer page, Integer size) {
        log.info("Searching products with query: {}, page: {}, size: {}", query, page, size);
        // Delegate to product service with search parameter
        return productService.listProducts(page, size, null, null, null, query);
    }
}

package org.spatino.apigateway.controller;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.api.SearchApi;
import org.spatino.apigateway.generated.model.ProductListResponse;
import org.spatino.apigateway.service.SearchService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

@Slf4j
@RestController
@RequiredArgsConstructor
public class SearchController implements SearchApi {

    private final SearchService searchService;

    @Override
    public Mono<ResponseEntity<ProductListResponse>> searchProducts(String q, Integer page, Integer size, ServerWebExchange exchange) {
        return searchService.searchProducts(q, page, size)
                .map(response -> ResponseEntity.ok()
                        .header("X-RateLimit-Limit", "100")
                        .header("X-RateLimit-Remaining", "95")
                        .header("X-Cache-Status", "MISS")
                        .body(response));
    }
}

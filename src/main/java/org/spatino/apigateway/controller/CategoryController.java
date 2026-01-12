package org.spatino.apigateway.controller;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.api.CategoriesApi;
import org.spatino.apigateway.generated.model.CategoryListResponse;
import org.spatino.apigateway.service.CategoryService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

@Slf4j
@RestController
@RequiredArgsConstructor
public class CategoryController implements CategoriesApi {

    private final CategoryService categoryService;

    @Override
    public Mono<ResponseEntity<CategoryListResponse>> listCategories(ServerWebExchange exchange) {
        return categoryService.listCategories()
                .map(response -> ResponseEntity.ok()
                        .header("X-Cache-Status", "MISS")
                        .body(response));
    }
}

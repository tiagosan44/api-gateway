package org.spatino.apigateway.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.model.Category;
import org.spatino.apigateway.generated.model.CategoryListResponse;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
public class CategoryService {

    private final ProductService productService;
    private final Map<String, Category> categoryStore = new ConcurrentHashMap<>();

    public CategoryService(ProductService productService) {
        this.productService = productService;
        initializeCategories();
    }

    private void initializeCategories() {
        createCategory("electronics", "Electronics", "Electronic devices and accessories");
        createCategory("home", "Home", "Home appliances and furniture");
        createCategory("sports", "Sports", "Sports equipment and apparel");
        createCategory("furniture", "Furniture", "Office and home furniture");
    }

    private void createCategory(String id, String name, String description) {
        Category category = new Category();
        category.setId(id);
        category.setName(name);
        category.setDescription(description);
        category.setProductCount(0);
        categoryStore.put(id, category);
    }

    public Mono<CategoryListResponse> listCategories() {
        log.info("Listing all categories");
        return Mono.fromCallable(() -> {
            List<Category> categories = new ArrayList<>(categoryStore.values());
            
            // Update product counts dynamically
            categories.forEach(category -> {
                long count = productService.listProducts(0, Integer.MAX_VALUE, category.getName(), null, null, null)
                        .map(response -> response.getProducts().size())
                        .block();
                category.setProductCount((int) count);
            });

            CategoryListResponse response = new CategoryListResponse();
            response.setCategories(categories);
            return response;
        });
    }
}

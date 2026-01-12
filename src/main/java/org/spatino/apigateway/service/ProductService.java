package org.spatino.apigateway.service;

import lombok.extern.slf4j.Slf4j;
import org.spatino.apigateway.generated.model.Product;
import org.spatino.apigateway.generated.model.ProductListResponse;
import org.spatino.apigateway.generated.model.ProductUpdate;
import org.spatino.apigateway.generated.model.Pagination;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.math.BigDecimal;
import java.net.URI;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
public class ProductService {

    // In-memory storage simulating a database
    private final Map<UUID, Product> productStore = new ConcurrentHashMap<>();

    public ProductService() {
        // Initialize with some sample data
        initializeSampleData();
    }

    private void initializeSampleData() {
        createSampleProduct("Laptop", "High-performance laptop", new BigDecimal("1299.99"), "Electronics", "TechBrand");
        createSampleProduct("Smartphone", "Latest smartphone model", new BigDecimal("899.99"), "Electronics", "PhoneBrand");
        createSampleProduct("Coffee Maker", "Automatic coffee maker", new BigDecimal("89.99"), "Home", "HomeBrand");
        createSampleProduct("Running Shoes", "Comfortable running shoes", new BigDecimal("129.99"), "Sports", "SportBrand");
        createSampleProduct("Desk Chair", "Ergonomic office chair", new BigDecimal("299.99"), "Furniture", "OfficeBrand");
    }

    private void createSampleProduct(String name, String description, BigDecimal price, String category, String brand) {
        Product product = new Product();
        product.setId(UUID.randomUUID());
        product.setName(name);
        product.setDescription(description);
        product.setPrice(price);
        product.setCategory(category);
        product.setBrand(brand);
        product.setImageUrl(URI.create("https://example.com/images/" + name.toLowerCase().replace(" ", "-") + ".jpg"));
        product.setCreatedAt(OffsetDateTime.now());
        product.setUpdatedAt(OffsetDateTime.now());
        productStore.put(product.getId(), product);
    }

    public Mono<ProductListResponse> listProducts(Integer page, Integer size, String category,
                                                   BigDecimal minPrice, BigDecimal maxPrice, String search) {
        log.info("Listing products: page={}, size={}, category={}, minPrice={}, maxPrice={}, search={}",
                page, size, category, minPrice, maxPrice, search);

        return Mono.fromCallable(() -> {
            List<Product> allProducts = new ArrayList<>(productStore.values());

            // Apply filters
            List<Product> filteredProducts = allProducts.stream()
                    .filter(p -> category == null || category.equals(p.getCategory()))
                    .filter(p -> minPrice == null || p.getPrice().compareTo(minPrice) >= 0)
                    .filter(p -> maxPrice == null || p.getPrice().compareTo(maxPrice) <= 0)
                    .filter(p -> search == null || p.getName().toLowerCase().contains(search.toLowerCase())
                            || (p.getDescription() != null && p.getDescription().toLowerCase().contains(search.toLowerCase())))
                    .toList();

            // Pagination
            int totalElements = filteredProducts.size();
            int totalPages = (int) Math.ceil((double) totalElements / size);
            int fromIndex = page * size;
            int toIndex = Math.min(fromIndex + size, totalElements);

            List<Product> pageProducts = fromIndex < totalElements ?
                    filteredProducts.subList(fromIndex, toIndex) : new ArrayList<>();

            Pagination pagination = new Pagination();
            pagination.setPage(page);
            pagination.setSize(size);
            pagination.setTotalElements(totalElements);
            pagination.setTotalPages(totalPages);

            ProductListResponse response = new ProductListResponse();
            response.setProducts(pageProducts);
            response.setPagination(pagination);

            return response;
        });
    }

    public Mono<Product> getProduct(UUID productId) {
        log.info("Getting product: {}", productId);
        return Mono.justOrEmpty(productStore.get(productId));
    }

    public Mono<Product> updateProduct(UUID productId, ProductUpdate productUpdate) {
        log.info("Updating product: {}", productId);
        return Mono.fromCallable(() -> {
            Product product = productStore.get(productId);
            if (product == null) {
                return null;
            }

            if (productUpdate.getName() != null) {
                product.setName(productUpdate.getName());
            }
            if (productUpdate.getDescription() != null) {
                product.setDescription(productUpdate.getDescription());
            }
            if (productUpdate.getPrice() != null) {
                product.setPrice(productUpdate.getPrice());
            }
            if (productUpdate.getCategory() != null) {
                product.setCategory(productUpdate.getCategory());
            }
            if (productUpdate.getBrand() != null) {
                product.setBrand(productUpdate.getBrand());
            }
            if (productUpdate.getImageUrl() != null) {
                product.setImageUrl(productUpdate.getImageUrl());
            }
            product.setUpdatedAt(OffsetDateTime.now());

            return product;
        });
    }

    public Mono<Boolean> deleteProduct(UUID productId) {
        log.info("Deleting product: {}", productId);
        return Mono.fromCallable(() -> productStore.remove(productId) != null);
    }
}

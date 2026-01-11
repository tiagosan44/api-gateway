# Code Generation from OpenAPI Specs

This project uses **OpenAPI Generator** to automatically generate code from specifications, following the Spec-Driven Development methodology.

## 🎯 What Gets Generated

### 1. API Interfaces (Server)
- **Location**: `target/generated-sources/openapi/src/main/java/org/spatino/apigateway/generated/api/`
- **Purpose**: Spring WebFlux reactive API interfaces
- **Features**:
  - Reactive interfaces using `Mono<>` and `Flux<>`
  - Bean validation annotations
  - Request/Response mappings
  - Spring Boot 4.0 compatible

**Generated APIs**:
- `ProductsApi.java` - Product CRUD operations
- `InventoryApi.java` - Inventory management
- `CategoriesApi.java` - Category listing
- `SearchApi.java` - Product search
- `MonitoringApi.java` - Health checks

### 2. Data Models (DTOs)
- **Location**: `target/generated-sources/openapi/src/main/java/org/spatino/apigateway/generated/model/`
- **Purpose**: Request/Response DTOs with validation
- **Features**:
  - Jakarta Bean Validation
  - Serializable models
  - Immutable where possible

**Generated Models**:
- `Product.java`
- `ProductUpdate.java`
- `ProductListResponse.java`
- `Inventory.java`
- `Category.java`
- `HealthResponse.java`
- And more...

### 3. Client (WebClient)
- **Location**: `target/generated-sources/openapi-client/src/main/java/org/spatino/apigateway/generated/client/`
- **Purpose**: Java WebClient for downstream services
- **Features**:
  - Reactive WebClient implementation
  - Type-safe API calls
  - Auto-generated test files

## 🔧 How to Generate Code

### Manual Generation
```bash
# Generate all code from specs
mvn generate-sources

# Clean and regenerate
mvn clean generate-sources
```

### Automatic Generation
Code generation runs automatically during:
- `mvn compile`
- `mvn package`
- `mvn install`

## 📋 Maven Plugin Configuration

The OpenAPI Generator is configured in `pom.xml` with two executions:

### Server Generation
```xml
<execution>
    <id>generate-api-server</id>
    <generatorName>spring</generatorName>
    <library>spring-boot</library>
    <configOptions>
        <useSpringBoot3>true</useSpringBoot3>
        <reactive>true</reactive>
        <interfaceOnly>true</interfaceOnly>
        <useJakartaEe>true</useJakartaEe>
    </configOptions>
</execution>
```

### Client Generation
```xml
<execution>
    <id>generate-api-client</id>
    <generatorName>java</generatorName>
    <library>webclient</library>
</execution>
```

## 🚫 Important Rules

### DO NOT Modify Generated Files
- Generated files are **read-only**
- Regenerated on every build
- Changes will be lost

### DO Implement Interfaces
Create your own implementations:

```java
@RestController
public class ProductsController implements ProductsApi {

    @Override
    public Mono<ResponseEntity<ProductListResponse>> listProducts(
            Integer page,
            Integer size,
            String category,
            BigDecimal minPrice,
            BigDecimal maxPrice,
            String search,
            ServerWebExchange exchange) {
        // Your implementation here
    }
}
```

## 🔄 CI/CD Integration

### GitHub Actions Workflows

#### 1. Spec Validation (`.github/workflows/spec-validation.yml`)
- Validates OpenAPI spec syntax
- Enforces required spec files
- Validates spec structure
- **Trigger**: Pull requests

#### 2. Code Generation Validation (`.github/workflows/code-generation-validation.yml`)
- Generates code from specs
- Verifies all expected files are created
- Validates package structure
- Counts generated files
- **Trigger**: Push and pull requests

#### 3. Build and Test (`.github/workflows/build-and-test.yml`)
- Generates code
- Compiles project
- Runs tests
- Packages application
- Uploads artifacts
- **Trigger**: Push to main, develop, feature branches

## 📊 Verification

After generation, verify with:

```bash
# Count generated API files
find target/generated-sources/openapi -name "*Api.java" | wc -l
# Expected: 5

# Count generated model files
find target/generated-sources/openapi -name "*.java" -path "*/model/*" | wc -l
# Expected: 14+

# List all generated APIs
find target/generated-sources/openapi -name "*Api.java"
```

## 🔗 Spec Location

All specifications are in the `specs/` directory:
- `specs/openapi.yaml` - API contract (source of truth)
- `specs/rate-limiting.md` - Rate limiting behavior
- `specs/load-balancing.md` - Load balancing strategy
- `specs/caching.md` - Caching policy
- `specs/observability.md` - Monitoring & logging

## 🐛 Troubleshooting

### Issue: Generation Fails
```bash
# Check OpenAPI spec is valid
mvn clean
mvn org.openapitools:openapi-generator-maven-plugin:7.12.0:generate
```

### Issue: Changes Not Reflected
```bash
# Clean all generated code and rebuild
mvn clean generate-sources
```

### Issue: Compilation Errors
- Ensure Java 25 is installed
- Check Spring Boot 4.0.1 compatibility
- Verify all dependencies are downloaded

## 📚 Resources

- [OpenAPI Generator Documentation](https://openapi-generator.tech/)
- [Spring Generator Options](https://openapi-generator.tech/docs/generators/spring)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest.html)

---

## ✅ Next Steps

Now that code generation is configured:

1. **Implement Controllers** - Create classes that implement generated API interfaces
2. **Add Business Logic** - Implement services behind the contract
3. **Write Tests** - Add contract tests to verify implementation matches spec
4. **Deploy** - Use CI/CD pipeline to validate and deploy

See `SDD-practical-guide.md` step 5 for implementation guidance.

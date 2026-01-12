package org.spatino.apigateway.filter;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ServerWebExchange;
import org.springframework.web.server.WebFilter;
import org.springframework.web.server.WebFilterChain;
import reactor.core.publisher.Mono;

import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Slf4j
@Component
public class RateLimitingFilter implements WebFilter {

    private static final int RATE_LIMIT = 100; // requests per window
    private static final Duration WINDOW_DURATION = Duration.ofMinutes(1);
    
    // In-memory storage for rate limiting (client IP -> request count)
    private final Map<String, RateLimitInfo> rateLimitStore = new ConcurrentHashMap<>();
    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public Mono<Void> filter(ServerWebExchange exchange, WebFilterChain chain) {
        String clientIp = getClientIp(exchange);
        
        RateLimitInfo info = rateLimitStore.computeIfAbsent(clientIp, k -> new RateLimitInfo());
        
        // Clean up expired entries
        if (info.isExpired()) {
            info.reset();
        }
        
        // Add rate limit headers
        exchange.getResponse().getHeaders().add("X-RateLimit-Limit", String.valueOf(RATE_LIMIT));
        exchange.getResponse().getHeaders().add("X-RateLimit-Remaining", 
                String.valueOf(Math.max(0, RATE_LIMIT - info.getCount())));
        exchange.getResponse().getHeaders().add("X-RateLimit-Reset", 
                String.valueOf(info.getResetTime()));
        
        // Check if rate limit exceeded
        if (info.getCount() >= RATE_LIMIT) {
            log.warn("Rate limit exceeded for client: {}", clientIp);
            return handleRateLimitExceeded(exchange, info);
        }
        
        // Increment counter
        info.increment();
        
        return chain.filter(exchange);
    }

    private String getClientIp(ServerWebExchange exchange) {
        String xForwardedFor = exchange.getRequest().getHeaders().getFirst("X-Forwarded-For");
        if (xForwardedFor != null && !xForwardedFor.isEmpty()) {
            return xForwardedFor.split(",")[0].trim();
        }
        
        if (exchange.getRequest().getRemoteAddress() != null) {
            return exchange.getRequest().getRemoteAddress().getAddress().getHostAddress();
        }
        
        return "unknown";
    }

    private Mono<Void> handleRateLimitExceeded(ServerWebExchange exchange, RateLimitInfo info) {
        exchange.getResponse().setStatusCode(HttpStatus.TOO_MANY_REQUESTS);
        exchange.getResponse().getHeaders().setContentType(MediaType.APPLICATION_JSON);
        
        long retryAfter = (info.getResetTime() - System.currentTimeMillis() / 1000);
        
        Map<String, Object> errorResponse = Map.of(
                "error", "Rate limit exceeded",
                "retryAfter", retryAfter
        );
        
        try {
            byte[] bytes = objectMapper.writeValueAsBytes(errorResponse);
            DataBuffer buffer = exchange.getResponse().bufferFactory().wrap(bytes);
            return exchange.getResponse().writeWith(Mono.just(buffer));
        } catch (JsonProcessingException e) {
            log.error("Error creating rate limit response", e);
            return exchange.getResponse().setComplete();
        }
    }

    private static class RateLimitInfo {
        private final AtomicInteger count = new AtomicInteger(0);
        private volatile long resetTime;

        public RateLimitInfo() {
            this.resetTime = System.currentTimeMillis() / 1000 + WINDOW_DURATION.getSeconds();
        }

        public int getCount() {
            return count.get();
        }

        public void increment() {
            count.incrementAndGet();
        }

        public long getResetTime() {
            return resetTime;
        }

        public boolean isExpired() {
            return System.currentTimeMillis() / 1000 > resetTime;
        }

        public void reset() {
            count.set(0);
            resetTime = System.currentTimeMillis() / 1000 + WINDOW_DURATION.getSeconds();
        }
    }
}

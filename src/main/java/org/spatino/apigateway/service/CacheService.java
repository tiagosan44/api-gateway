package org.spatino.apigateway.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.time.Duration;

@Slf4j
@Service
@RequiredArgsConstructor
public class CacheService {

    private final ReactiveRedisTemplate<String, Object> redisTemplate;
    private static final Duration DEFAULT_TTL = Duration.ofMinutes(5);

    public <T> Mono<T> get(String key, Class<T> type) {
        log.debug("Getting from cache: {}", key);
        return redisTemplate.opsForValue()
                .get(key)
                .cast(type)
                .doOnSuccess(value -> {
                    if (value != null) {
                        log.debug("Cache HIT for key: {}", key);
                    } else {
                        log.debug("Cache MISS for key: {}", key);
                    }
                })
                .onErrorResume(e -> {
                    log.error("Error getting from cache: {}", key, e);
                    return Mono.empty();
                });
    }

    public Mono<Boolean> set(String key, Object value) {
        return set(key, value, DEFAULT_TTL);
    }

    public Mono<Boolean> set(String key, Object value, Duration ttl) {
        log.debug("Setting cache: {} with TTL: {}", key, ttl);
        return redisTemplate.opsForValue()
                .set(key, value, ttl)
                .doOnSuccess(success -> log.debug("Cache SET successful for key: {}", key))
                .onErrorResume(e -> {
                    log.error("Error setting cache: {}", key, e);
                    return Mono.just(false);
                });
    }

    public Mono<Boolean> delete(String key) {
        log.debug("Deleting from cache: {}", key);
        return redisTemplate.delete(key)
                .map(count -> count > 0)
                .doOnSuccess(deleted -> log.debug("Cache DELETE for key: {} - {}", key, deleted))
                .onErrorResume(e -> {
                    log.error("Error deleting from cache: {}", key, e);
                    return Mono.just(false);
                });
    }

    public Mono<Boolean> deletePattern(String pattern) {
        log.debug("Deleting from cache with pattern: {}", pattern);
        return redisTemplate.keys(pattern)
                .flatMap(redisTemplate::delete)
                .reduce(0L, Long::sum)
                .map(count -> count > 0)
                .doOnSuccess(deleted -> log.debug("Cache DELETE pattern: {} - {}", pattern, deleted))
                .onErrorResume(e -> {
                    log.error("Error deleting pattern from cache: {}", pattern, e);
                    return Mono.just(false);
                });
    }
}

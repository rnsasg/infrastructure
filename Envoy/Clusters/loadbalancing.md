# Load balancing
Load balancing is a way of distributing traffic between multiple endpoints in a single upstream cluster. The reason for distributing traffic across numerous endpoints is to make the best use of the available resources.

To achieve the most efficient use of resources, Envoy provides different load-balancing strategies that can be separated into two groups: global load balancing and distributed load balancing. The difference is that we use a single control plane that decides traffic distribution between the endpoints in the global load balancing. Envoy determines how the load gets distributed (e.g., using active health checking, zone-aware routing, and load balancing policy).

One of the techniques for distributing load among multiple endpoints is consistent hashing. The server uses a part of the request to create a hash value to select an endpoint. In modulo hashing, the hash is considered to be a huge number. To get the endpoint index to send the request to, we take the hash modulo the number of available endpoints (index=hash % endpointCount). This approach works well if the number of endpoints is stable. However, if the endpoints are added or removed (i.e., they are unhealthy, we scale them up or down, etc.), most requests will end up on a different endpoint.

Consistent hashing is an approach where each endpoint gets assigned multiple values based on some property. Then, each request gets assigned to the endpoint that has the nearest hash value. The value of this approach is that when we add or remove endpoints, most requests will end up on the same endpoint they did before. This “stickiness” is helpful because it won’t disturb any caches the endpoints hold.

## Load balancing policy
Envoy uses one of the load-balancing policies to select an endpoint to send the traffic to. The load balancing policy is configurable and can be specified for each upstream cluster separately. Note that the load balancing is only performed across healthy endpoints. If no active or passive health checking is defined, all endpoints are assumed to be healthy.

We can configure the load balancing policy using the lb_policy field and other fields specific to the selected policy.

### Weighted round-robin (default)
Weighted round-robin (ROUND_ROBIN) selects the endpoint in round-robin order. If endpoints are weighted, then a weighted round-robin schedule is used. This strategy gives us a predictable distribution of requests across all endpoints. The higher weighted endpoints will appear more often in the rotation to achieve effective weighting.

### Weighted least request
The weighted least request (LEAST_REQUEST) algorithm depends on the weights assigned to endpoints.

**If all endpoint weights are equal**, the algorithm selects N random available endpoints (choice_count) and picks the one with the fewest active requests.

**If endpoint weights are not equal**, the algorithm shifts into a mode where it uses a weighted round-robin schedule in which weights are dynamically adjusted based on the endpoint’s request load at the time of selection.

The following formula is used to calculate the weights dynamically:

> weight = load_balancing_weight / (active_requests + 1)^active_request_bias

The active_request_bias is configurable (defaults to 1.0). The larger the active request bias, the more aggressively active requests will lower the effective weight.

If active_request_bias is set to 0, the algorithm behaves like the round-robin and ignores the active request count at the picking time.

We can set the optional configuration for the weighted least request using the least_request_lb_config field:

```yaml
...
  lb_policy: LEAST_REQUEST
  least_request_lb_config:
    choice_count: 5
    active_request_bias: 0.5
...
```
## Ring hash
The ring hash (or modulo hash) algorithm (RING_HASH) implements consistent hashing to endpoints. Each endpoint address (default setting) is hashed and mapped onto a ring. Envoy routes the requests to an endpoint by hashing some request property and finding the nearest corresponding endpoint clockwise around the ring. The hash key defaults to the endpoint address; however, it can be changed to any other property using the hash_key field.

We can configure the ring hash algorithm by specifying the minimum (minimum_ring_size) and maximum ring size (maximum_ring_size) and use the stats (min_hashes_per_host and max_hashes_per_host) to ensure good distribution. The larger the ring, the better the request distribution will reflect the desired weights. The minimum ring size defaults to 1024 entries (limited to 8M entries), while the maximum ring size defaults to 8M (limited to 8M).

We can set the optional configuration for the ring hash can be set using the ring_hash_lb_config field:

```yaml
...
  lb_policy: RING_HASH
  ring_hash_lb_config:
    minimum_ring_size: 2000
    maximum_ring_size: 10000
...
```

## Maglev
Like the ring hash algorithm, the maglev (MAGLEV) algorithm also implements consistent hashing to endpoints. The algorithm produces a lookup table that allows finding an item within a constant time. Maglev was designed to be faster than the ring hash algorithm for the lookups and use less memory. You can read more about it in the article Maglev: [A Fast and Reliable Software Network Load Balancer](https://dgryski.medium.com/consistent-hashing-algorithmic-tradeoffs-ef6b8e2fcae8).

We can set the optional configuration for the maglev algorithm using the maglev_lb_config field:

```yaml
...
  lb_policy: MAGLEV
  maglev_lb_config:
    table_size: 69997
...
```

The default table size is 65537, but it can be set to any prime number if it’s not greater than 5000011.

## Original destination
The original destination is a special-purpose load balancer that can only be used with an original destination cluster. We mentioned the original destination load balancer when discussing the original destination cluster type.

## Random
As the name suggests, the random (RANDOM) algorithm picks an available random endpoint. If you don’t have an active health-checking policy configured, the random algorithm performs better than the round-robin.

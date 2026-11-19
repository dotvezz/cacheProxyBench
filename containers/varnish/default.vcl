vcl 4.1;

backend default {
    .host = "mockorigin";
    .port = "80";
}

sub vcl_recv {
    if (req.method != "GET" && req.method != "HEAD") {
        return (pass);
    }
    return (hash);
}

sub vcl_backend_response {
    if (beresp.http.Cache-Control ~ "no-cache|no-store|private" || beresp.status >= 500) {
        set beresp.uncacheable = true;
        set beresp.ttl = 0s;
        return (deliver);
    }
    set beresp.ttl = 5m;
    set beresp.grace = 1m;
    return (deliver);
}
sub vcl_deliver {
    if (req.method != "GET" && req.method != "HEAD") {
        set resp.http.Cache-Status = "Varnish; fwd=method";
    } else if (obj.uncacheable) {
        set resp.http.Cache-Status = "Varnish; fwd=bypass";
    } else if (resp.http.X-Fetched) {
        # This thread performed the origin fetch (the coalescing leader)
        unset resp.http.X-Fetched;
        set resp.http.Cache-Status = "Varnish; fwd=uri-miss; stored";
    } else if (obj.hits > 0 && req.restarts == 0 && resp.http.Age ~ "^[01]?$") {
        # Served from cache with near-zero age without fetching itself (coalesced follower)
        set resp.http.Cache-Status = "Varnish; hit; collapsed; ttl=" + obj.ttl;
    } else if (obj.hits > 0) {
        # Standard cache hit
        set resp.http.Cache-Status = "Varnish; hit; ttl=" + obj.ttl;
    } else {
        set resp.http.Cache-Status = "Varnish; fwd=uri-miss";
    }
}
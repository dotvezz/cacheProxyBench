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
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
    } else {
        set resp.http.X-Cache = "MISS";
    }
}

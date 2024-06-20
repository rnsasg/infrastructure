## Path matching

We’ve only discussed one matching rule that matches a prefix using the prefix field. The table below explains the other supported matching rules.

| Rule name        | Description                                                                                                  |
|------------------|--------------------------------------------------------------------------------------------------------------|
| prefix           | The prefix must match the beginning of the path header. For example, the prefix /api will match paths /api and /api/v1, but not /. |
| path             | The path must match the exact path header (without query string). For example, path /api will match paths /api, but not /api/v1 or /. |
| safe_regex       | The path must match the specified regular expression. For example, regex ^/products/\d+$ will match path /products/123 or /products/321 but not /products/hello or /api/products/123 |
| connect_matcher  | Matcher only matches CONNECT requests (currently in Alpha)                                                   |

By default, the prefix and path matching is case-sensitive. To make it case insensitive, we can set the case_sensitive to false. Note that this setting doesn’t apply to safe_regex matching.

## Headers matching
Another way of matching requests is by specifying a set of headers. The router checks the request headers against all specified headers in the route config. The match is made if all specified headers exist on the request and have the same values.

Multiple matching rules can be applied to headers.

## Range match
The range_match checks if the request header value is within the specified integer range in the base ten notation. The value can include an optional plus or minus sign followed by digits.

To use range matching, we specify the range's start and end. The start value is inclusive, while the end of the range is exclusive ([start, end)).

```yaml
- match:
    prefix: "/"
    headers:
    - name: minor_version
      range_match:
        start: 1
        end: 11
```

The above range match will match the minor_version header values if it’s set to any number between 1 and 10.

Present match

The present_match checks for a presence of a specific header in the incoming request.

```yaml
- match:
    prefix: "/"
    headers:
    - name: debug
      present_match: true
```

The above snippet will evaluate  true if we set the debug header, regardless of the header value. If we set the present_match value to false, we can check for the absence of the header.

## String match

The string_match allows us to match the exact header values by prefix or suffix, using regular expression or checking if the value contains a specific string.

```yaml
- match:
    prefix: "/"
    headers:
    # Header `regex_match` matches the provided regular expression
    - name: regex_match
      string_match:
        safe_regex_match:
          google_re2: {}
          regex: "^v\\d+$"
    # Header `exact_match` contains the value `hello`
    - name: exact_match
      string_match:
        exact: "hello"
    # Header `prefix_match` starts with `api`
    - name: prefix_match
      string_match:
        prefix: "api"
    # Header `suffix_match` ends with `_1`
    - name: suffix_match
      string_match:
        suffix: "_1"
    # Header `contains_match` contains the value "debug"
    - name: contains_match
      string_match:
        contains: "debug"
```

Invert match

If we set the invert_match, the match result is inverted.

```yaml
- match:
    prefix: "/"
    headers:
    - name: version
      range_match: 
        start: 1
        end: 6
      invert_match: true
```

The above snippet will check that the value in the version header falls between 1 and 5; however, because we added the invert_match field, it inverts the result and checks if the header values fall out of that range.

The invert_match can be used by other matchers. For example:

```yaml
- match:
    prefix: "/"
    headers:
    - name: env
      contains_match: "test"
      invert_match: true
```

The above snippet will check that the env header value doesn’t contain the string test. If we set the env header that doesn’t include the string test, the whole match evaluates to true.

Query parameters matching
Using the query_parameters field, we can specify the parameters from the URL query on which the route should match. The filter will check the query string from the path header and compare it against the provided parameters.

If more than one query parameter is specified, they must match the rule to evaluate to true.

Consider the following example:

```yaml
- match:
    prefix: "/"
    query_parameters:
    - name: env
      present_match: true
```

The above snippet will evaluate to true if there’s a query parameter called env set. It’s not saying anything about the value. It’s just checking for its presence. For example, the following request would evaluate to true using the above matcher:

GET /hello?env=test
We can also use a string matcher to check for the values of query parameters. The table below lists different rules for string matching.

| Rule name   | Description                                                                            |
|-------------|----------------------------------------------------------------------------------------|
| exact       | Must match the exact value of the query parameter.                                     |
| prefix      | The prefix must match the beginning of the query parameter value.                      |
| suffix      | The suffix must match the ending of the query parameter value.                         |
| safe_regex  | The query parameter value must match the specified regular expression.                 |
| contains    | Checks if the query parameter value contains a specific string.                        |

Here’s another example of a case-insensitive query parameter matching using the prefix rule:

```yaml
- match:
    prefix: "/"
    query_parameters:
    - name: env
      string_match:
        prefix: "env_"
        ignore_case: true
```

The above will evaluate to true if there’s a query parameter called env whose value starts with env_. For example, env_staging and ENV_prod evaluates to true.

gRPC and TLS matchers
We can configure the other two matchers on the routes: the gRPC route matcher (grpc) and TLS context matcher (tls_context).

The gRPC matcher will only match the gRPC requests. The router checks the content-type header for application/grpc and other application/grpc+ values to determine if the request is a gRPC request.

For example:

```yaml
- match:
    prefix: "/"
    grpc: {}
```

Note the gRPC matcher doesn’t have any options.

The above snippet will match the route if the request is a gRPC request.

Similarly, the TLS matcher, if specified, will match the TLS context against provided options. Within the tls_context field, we can define two boolean values – presented and validated. The presented field checks whether a certificate is presented or not. The validated field checks whether a certificate is validated or not.

For example:

```yaml
- match:
    prefix: "/"
    tls_context:
      presented: true
      validated: true
```
The above match evaluates to true if a certificate is presented and validated.



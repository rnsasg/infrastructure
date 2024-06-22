-- custom_filter.lua
function envoy_on_request(request_handle)
    request_handle:logInfo("Lua filter executed for request.")
    request_handle:headers():add("x-custom-header", "custom_value")
end

function envoy_on_response(response_handle)
    response_handle:logInfo("Lua filter executed for response.")
    response_handle:headers():add("x-response-header", "response_value")
end



# .CommonApi

All URIs are relative to *https://api.pdok.nl/bzk/location-api/autocomplete/v1-preprod*

| Method                                                                  | HTTP request         | Description                |
| ----------------------------------------------------------------------- | -------------------- | -------------------------- |
| [**getConformanceDeclaration**](CommonApi.md#getConformanceDeclaration) | **GET** /conformance | API conformance definition |
| [**getLandingPage**](CommonApi.md#getLandingPage)                       | **GET** /            | Landing page               |
| [**getOpenApi**](CommonApi.md#getOpenApi)                               | **GET** /api         | This document              |

# **getConformanceDeclaration**

> ConfClasses getConformanceDeclaration()

A list of all conformance classes specified in a standard that the server conforms to.

### Example

```typescript
import { createConfiguration, CommonApi } from ''
import type { CommonApiGetConformanceDeclarationRequest } from ''

const configuration = createConfiguration()
const apiInstance = new CommonApi(configuration)

const request: CommonApiGetConformanceDeclarationRequest = {
  // The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON. (optional)
  f: 'json',
}

const data = await apiInstance.getConformanceDeclaration(request)
console.log('API called successfully. Returned data:', data)
```

### Parameters

| Name  | Type                | Description                                                        | Notes                                                                                                                                            |
| ----- | ------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- |
| **f** | [\*\*&#39;json&#39; | &#39;html&#39;**]**Array<&#39;json&#39; &#124; &#39;html&#39;>\*\* | The optional f parameter indicates the output format that the server shall provide as part of the response document. The default format is JSON. | (optional) defaults to 'json' |

### Return type

**ConfClasses**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/html, application/problem+json

### HTTP response details

| Status code | Description                                                                                                                                                                                                                                                                                          | Response headers |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| **200**     | The URIs of all conformance classes supported by the server. To support \&quot;generic\&quot; clients that want to access multiple OGC API Features implementations - and not \&quot;just\&quot; a specific API / server, the server declares the conformance classes it implements and conforms to. | -                |
| **400**     | Bad request: For example, invalid or unknown query parameters.                                                                                                                                                                                                                                       | -                |
| **404**     | Not found: The requested resource does not exist on the server. For example, a path parameter had an incorrect value.                                                                                                                                                                                | -                |
| **406**     | Not acceptable: The requested media type is not supported by this resource.                                                                                                                                                                                                                          | -                |
| **500**     | Internal server error: An unexpected server error occurred.                                                                                                                                                                                                                                          | -                |
| **502**     | Bad Gateway: An unexpected error occurred while forwarding/proxying the request to another server.                                                                                                                                                                                                   | -                |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getLandingPage**

> LandingPage getLandingPage()

The landing page provides links to the API definition and the conformance statements for this API.

### Example

```typescript
import { createConfiguration, CommonApi } from ''
import type { CommonApiGetLandingPageRequest } from ''

const configuration = createConfiguration()
const apiInstance = new CommonApi(configuration)

const request: CommonApiGetLandingPageRequest = {
  // The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON. (optional)
  f: 'json',
}

const data = await apiInstance.getLandingPage(request)
console.log('API called successfully. Returned data:', data)
```

### Parameters

| Name  | Type                | Description                                                        | Notes                                                                                                                                            |
| ----- | ------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- |
| **f** | [\*\*&#39;json&#39; | &#39;html&#39;**]**Array<&#39;json&#39; &#124; &#39;html&#39;>\*\* | The optional f parameter indicates the output format that the server shall provide as part of the response document. The default format is JSON. | (optional) defaults to 'json' |

### Return type

**LandingPage**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, application/vnd.oai.openapi+json;version=3.0, text/html, application/problem+json

### HTTP response details

| Status code | Description                                                                                                                                                                                                                          | Response headers |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------- |
| **200**     | The landing page provides links to the API definition (link relations &#x60;service-desc&#x60; and &#x60;service-doc&#x60;), and the Conformance declaration (path &#x60;/conformance&#x60;, link relation &#x60;conformance&#x60;). | -                |
| **400**     | Bad request: For example, invalid or unknown query parameters.                                                                                                                                                                       | -                |
| **404**     | Not found: The requested resource does not exist on the server. For example, a path parameter had an incorrect value.                                                                                                                | -                |
| **406**     | Not acceptable: The requested media type is not supported by this resource.                                                                                                                                                          | -                |
| **500**     | Internal server error: An unexpected server error occurred.                                                                                                                                                                          | -                |
| **502**     | Bad Gateway: An unexpected error occurred while forwarding/proxying the request to another server.                                                                                                                                   | -                |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getOpenApi**

> void getOpenApi()

This document

### Example

```typescript
import { createConfiguration, CommonApi } from ''
import type { CommonApiGetOpenApiRequest } from ''

const configuration = createConfiguration()
const apiInstance = new CommonApi(configuration)

const request: CommonApiGetOpenApiRequest = {
  // The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON. (optional)
  f: 'json',
}

const data = await apiInstance.getOpenApi(request)
console.log('API called successfully. Returned data:', data)
```

### Parameters

| Name  | Type                | Description                                                        | Notes                                                                                                                                            |
| ----- | ------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- |
| **f** | [\*\*&#39;json&#39; | &#39;html&#39;**]**Array<&#39;json&#39; &#124; &#39;html&#39;>\*\* | The optional f parameter indicates the output format that the server shall provide as part of the response document. The default format is JSON. | (optional) defaults to 'json' |

### Return type

**void**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/problem+json

### HTTP response details

| Status code | Description                                                                                                           | Response headers |
| ----------- | --------------------------------------------------------------------------------------------------------------------- | ---------------- |
| **200**     | General Success response.                                                                                             | -                |
| **400**     | Bad request: For example, invalid or unknown query parameters.                                                        | -                |
| **404**     | Not found: The requested resource does not exist on the server. For example, a path parameter had an incorrect value. | -                |
| **406**     | Not acceptable: The requested media type is not supported by this resource.                                           | -                |
| **500**     | Internal server error: An unexpected server error occurred.                                                           | -                |
| **502**     | Bad Gateway: An unexpected error occurred while forwarding/proxying the request to another server.                    | -                |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# .CollectionsApi

All URIs are relative to *https://api.pdok.nl/bzk/location-api/autocomplete/v1-preprod*

| Method                                                 | HTTP request         | Description                    |
| ------------------------------------------------------ | -------------------- | ------------------------------ |
| [**getCollections**](CollectionsApi.md#getCollections) | **GET** /collections | the collections in the dataset |

# **getCollections**

> Collections getCollections()

A list of all collections (geospatial data resources) in this dataset.

### Example

```typescript
import { createConfiguration, CollectionsApi } from ''
import type { CollectionsApiGetCollectionsRequest } from ''

const configuration = createConfiguration()
const apiInstance = new CollectionsApi(configuration)

const request: CollectionsApiGetCollectionsRequest = {
  // The format of the response. If no value is provided, the standard http rules apply, i.e., the accept header is used to determine the format.  Pre-defined values are \"json\" and \"html\". The response to other values is determined by the server. (optional)
  f: 'json',
}

const data = await apiInstance.getCollections(request)
console.log('API called successfully. Returned data:', data)
```

### Parameters

| Name  | Type                | Description                                                        | Notes                                                                                                                                                                                                                                                                    |
| ----- | ------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------- |
| **f** | [\*\*&#39;json&#39; | &#39;html&#39;**]**Array<&#39;json&#39; &#124; &#39;html&#39;>\*\* | The format of the response. If no value is provided, the standard http rules apply, i.e., the accept header is used to determine the format. Pre-defined values are \&quot;json\&quot; and \&quot;html\&quot;. The response to other values is determined by the server. | (optional) defaults to undefined |

### Return type

**Collections**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/html, application/problem+json

### HTTP response details

| Status code | Description                                                                                                                                                                                                                                                                                                                                                                          | Response headers |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------- |
| **200**     | The collections (geospatial data resources) shared by this API. This response can be references directly for every service that wants only essential information at the collections level. /collections/collectionId might return more information. The dataset is organized as one or more collections. This resource provides information about and how to access the collections. | -                |
| **400**     | Bad request: For example, invalid or unknown query parameters.                                                                                                                                                                                                                                                                                                                       | -                |
| **404**     | Not found: The requested resource does not exist on the server. For example, a path parameter had an incorrect value.                                                                                                                                                                                                                                                                | -                |
| **406**     | Not acceptable: The requested media type is not supported by this resource.                                                                                                                                                                                                                                                                                                          | -                |
| **500**     | Internal server error: An unexpected server error occurred.                                                                                                                                                                                                                                                                                                                          | -                |
| **502**     | Bad Gateway: An unexpected error occurred while forwarding/proxying the request to another server.                                                                                                                                                                                                                                                                                   | -                |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

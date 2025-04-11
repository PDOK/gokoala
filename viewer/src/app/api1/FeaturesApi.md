# .FeaturesApi

All URIs are relative to *https://api.pdok.nl/bzk/location-api/autocomplete/v1-preprod*

| Method                              | HTTP request    | Description                                                 |
| ----------------------------------- | --------------- | ----------------------------------------------------------- |
| [**search**](FeaturesApi.md#search) | **GET** /search | search features in one or more collections across datasets. |

# **search**

> FeatureCollectionGeoJSON search()

This endpoint allows one to implement autocomplete functionality for location search. The `q` parameter accepts a partial location name and will return all matching locations up to the specified `limit`. The list of search results are offered as features (in GeoJSON, JSON-FG) but contain only minimal information; like a feature ID, highlighted text and a bounding box. When you want to get the full feature you must follow the included link (`href`) in the search result. This allows one to retrieve all properties of the feature and the full geometry from the corresponding OGC API.

### Example

```typescript
import { createConfiguration, FeaturesApi } from ''
import type { FeaturesApiSearchRequest } from ''

console.log('API called successfully. Returned data:', data)
```

### Parameters

| Name                  | Type                                                        | Description                                                                                                                                                                                                                                                                                                                                                                                                 | Notes                                                                                                  |
| --------------------- | ----------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------- |
| **q**                 | [**string**]                                                | The search term(s)                                                                                                                                                                                                                                                                                                                                                                                          | defaults to undefined                                                                                  |
| **functioneelGebied** | **SearchFunctioneelGebiedParameter**                        | When provided the functioneel_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the functioneel_gebied collection, for example &#x60;q&#x3D;foo&amp;functioneel_gebied[version]&#x3D;1&amp;functioneel_gebied[relevance]&#x3D;0.5&#x60; | (optional) defaults to undefined                                                                       |
| **geografischGebied** | **SearchGeografischGebiedParameter**                        | When provided the geografisch_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the geografisch_gebied collection, for example &#x60;q&#x3D;foo&amp;geografisch_gebied[version]&#x3D;1&amp;geografisch_gebied[relevance]&#x3D;0.5&#x60; | (optional) defaults to undefined                                                                       |
| **ligplaats**         | **SearchLigplaatsParameter**                                | When provided the ligplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the ligplaats collection, for example &#x60;q&#x3D;foo&amp;ligplaats[version]&#x3D;1&amp;ligplaats[relevance]&#x3D;0.5&#x60;                                     | (optional) defaults to undefined                                                                       |
| **standplaats**       | **SearchStandplaatsParameter**                              | When provided the standplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the standplaats collection, for example &#x60;q&#x3D;foo&amp;standplaats[version]&#x3D;1&amp;standplaats[relevance]&#x3D;0.5&#x60;                             | (optional) defaults to undefined                                                                       |
| **verblijfsobject**   | **SearchVerblijfsobjectParameter**                          | When provided the verblijfsobject collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the verblijfsobject collection, for example &#x60;q&#x3D;foo&amp;verblijfsobject[version]&#x3D;1&amp;verblijfsobject[relevance]&#x3D;0.5&#x60;             | (optional) defaults to undefined                                                                       |
| **woonplaats**        | **SearchWoonplaatsParameter**                               | When provided the woonplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the woonplaats collection, for example &#x60;q&#x3D;foo&amp;woonplaats[version]&#x3D;1&amp;woonplaats[relevance]&#x3D;0.5&#x60;                                 | (optional) defaults to undefined                                                                       |
| **limit**             | [**number**]                                                | The optional limit parameter limits the number of items that are presented in the response document. Only items are counted that are on the first level of the collection in the response document. Nested objects contained within the explicitly requested items shall not be counted. Minimum &#x3D; 1. Maximum &#x3D; 50. Default &#x3D; 10.                                                            | (optional) defaults to 10                                                                              |
| **crs**               | [\*\*&#39;http://www.opengis.net/def/crs/OGC/1.3/CRS84&#39; | &#39;http://www.opengis.net/def/crs/EPSG/0/28992&#39;**]**Array<&#39;http://www.opengis.net/def/crs/OGC/1.3/CRS84&#39; &#124; &#39;http://www.opengis.net/def/crs/EPSG/0/28992&#39;>**                                                                                                                                                                                                                      | The coordinate reference system of the geometries in the response. Default is WGS84 longitude/latitude | (optional) defaults to 'http://www.opengis.net/def/crs/OGC/1.3/CRS84' |

### Return type

**FeatureCollectionGeoJSON**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/geo+json, application/vnd.ogc.fg+json, text/html, application/problem+json

### HTTP response details

| Status code | Description                                                                                                                                                                                                                           | Response headers                                                                                                            |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| **200**     | The response is a document consisting of features in the collection. The features contain only minimal information but include a link (href) to the actual feature in another OGC API. Follow that link to get the full feature data. | \* Content-Crs - a URI, in angular brackets, identifying the coordinate reference system used in the content / payload <br> |
| **400**     | Bad request: For example, invalid or unknown query parameters.                                                                                                                                                                        | -                                                                                                                           |
| **404**     | Not found: The requested resource does not exist on the server. For example, a path parameter had an incorrect value.                                                                                                                 | -                                                                                                                           |
| **406**     | Not acceptable: The requested media type is not supported by this resource.                                                                                                                                                           | -                                                                                                                           |
| **500**     | Internal server error: An unexpected server error occurred.                                                                                                                                                                           | -                                                                                                                           |
| **502**     | Bad Gateway: An unexpected error occurred while forwarding/proxying the request to another server.                                                                                                                                    | -                                                                                                                           |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

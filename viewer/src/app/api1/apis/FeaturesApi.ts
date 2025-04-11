// TODO: better import syntax?
import { BaseAPIRequestFactory, RequiredError, COLLECTION_FORMATS } from './baseapi'
import { Configuration } from '../configuration'
import { RequestContext, HttpMethod, ResponseContext, HttpFile, HttpInfo } from '../http/http'
import { ObjectSerializer } from '../models/ObjectSerializer'
import { ApiException } from './exception'
import { canConsumeForm, isCodeInRange } from '../util'
import { SecurityAuthentication } from '../auth/auth'

import { Exception } from '../models/Exception'
import { FeatureCollectionGeoJSON } from '../models/FeatureCollectionGeoJSON'
import { FeatureCollectionJSONFG } from '../models/FeatureCollectionJSONFG'
import { SearchFunctioneelGebiedParameter } from '../models/SearchFunctioneelGebiedParameter'
import { SearchGeografischGebiedParameter } from '../models/SearchGeografischGebiedParameter'
import { SearchLigplaatsParameter } from '../models/SearchLigplaatsParameter'
import { SearchStandplaatsParameter } from '../models/SearchStandplaatsParameter'
import { SearchVerblijfsobjectParameter } from '../models/SearchVerblijfsobjectParameter'
import { SearchWoonplaatsParameter } from '../models/SearchWoonplaatsParameter'

/**
 * no description
 */
export class FeaturesApiRequestFactory extends BaseAPIRequestFactory {
  /**
   * This endpoint allows one to implement autocomplete functionality for location search. The `q` parameter accepts a partial location name and will return all matching locations up to the specified `limit`. The list of search results are offered as features (in GeoJSON, JSON-FG) but contain only minimal information; like a feature ID, highlighted text and a bounding box. When you want to get the full feature you must follow the included link (`href`) in the search result. This allows one to retrieve all properties of the feature and the full geometry from the corresponding OGC API.
   * search features in one or more collections across datasets.
   * @param q The search term(s)
   * @param functioneelGebied When provided the functioneel_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the functioneel_gebied collection, for example &#x60;q&#x3D;foo&amp;functioneel_gebied[version]&#x3D;1&amp;functioneel_gebied[relevance]&#x3D;0.5&#x60;
   * @param geografischGebied When provided the geografisch_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the geografisch_gebied collection, for example &#x60;q&#x3D;foo&amp;geografisch_gebied[version]&#x3D;1&amp;geografisch_gebied[relevance]&#x3D;0.5&#x60;
   * @param ligplaats When provided the ligplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the ligplaats collection, for example &#x60;q&#x3D;foo&amp;ligplaats[version]&#x3D;1&amp;ligplaats[relevance]&#x3D;0.5&#x60;
   * @param standplaats When provided the standplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the standplaats collection, for example &#x60;q&#x3D;foo&amp;standplaats[version]&#x3D;1&amp;standplaats[relevance]&#x3D;0.5&#x60;
   * @param verblijfsobject When provided the verblijfsobject collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the verblijfsobject collection, for example &#x60;q&#x3D;foo&amp;verblijfsobject[version]&#x3D;1&amp;verblijfsobject[relevance]&#x3D;0.5&#x60;
   * @param woonplaats When provided the woonplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the woonplaats collection, for example &#x60;q&#x3D;foo&amp;woonplaats[version]&#x3D;1&amp;woonplaats[relevance]&#x3D;0.5&#x60;
   * @param limit The optional limit parameter limits the number of items that are presented in the response document.  Only items are counted that are on the first level of the collection in the response document. Nested objects contained within the explicitly requested items shall not be counted.  Minimum &#x3D; 1. Maximum &#x3D; 50. Default &#x3D; 10.
   * @param crs The coordinate reference system of the geometries in the response. Default is WGS84 longitude/latitude
   */
  public async search(
    q: string,
    functioneelGebied?: SearchFunctioneelGebiedParameter,
    geografischGebied?: SearchGeografischGebiedParameter,
    ligplaats?: SearchLigplaatsParameter,
    standplaats?: SearchStandplaatsParameter,
    verblijfsobject?: SearchVerblijfsobjectParameter,
    woonplaats?: SearchWoonplaatsParameter,
    limit?: number,
    crs?: 'http://www.opengis.net/def/crs/OGC/1.3/CRS84' | 'http://www.opengis.net/def/crs/EPSG/0/28992',
    _options?: Configuration
  ): Promise<RequestContext> {
    let _config = _options || this.configuration

    // verify required parameter 'q' is not null or undefined
    if (q === null || q === undefined) {
      throw new RequiredError('FeaturesApi', 'search', 'q')
    }

    // Path Params
    const localVarPath = '/search'

    // Make Request Context
    const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET)
    requestContext.setHeaderParam('Accept', 'application/json, */*;q=0.8')

    // Query Params
    if (q !== undefined) {
      requestContext.setQueryParam('q', ObjectSerializer.serialize(q, 'string', ''))
    }

    // Query Params
    if (functioneelGebied !== undefined) {
      const serializedParams = ObjectSerializer.serialize(functioneelGebied, 'SearchFunctioneelGebiedParameter', '')
      for (const key of Object.keys(serializedParams)) {
        requestContext.setQueryParam(key, serializedParams[key])
      }
    }

    // Query Params
    if (geografischGebied !== undefined) {
      const serializedParams = ObjectSerializer.serialize(geografischGebied, 'SearchGeografischGebiedParameter', '')
      for (const key of Object.keys(serializedParams)) {
        requestContext.setQueryParam(key, serializedParams[key])
      }
    }

    // Query Params
    if (ligplaats !== undefined) {
      const serializedParams = ObjectSerializer.serialize(ligplaats, 'SearchLigplaatsParameter', '')
      for (const key of Object.keys(serializedParams)) {
        requestContext.setQueryParam(key, serializedParams[key])
      }
    }

    // Query Params
    if (standplaats !== undefined) {
      const serializedParams = ObjectSerializer.serialize(standplaats, 'SearchStandplaatsParameter', '')
      for (const key of Object.keys(serializedParams)) {
        requestContext.setQueryParam(key, serializedParams[key])
      }
    }

    // Query Params
    if (verblijfsobject !== undefined) {
      const serializedParams = ObjectSerializer.serialize(verblijfsobject, 'SearchVerblijfsobjectParameter', '')
      for (const key of Object.keys(serializedParams)) {
        requestContext.setQueryParam(key, serializedParams[key])
      }
    }

    // Query Params
    if (woonplaats !== undefined) {
      const serializedParams = ObjectSerializer.serialize(woonplaats, 'SearchWoonplaatsParameter', '')
      for (const key of Object.keys(serializedParams)) {
        requestContext.setQueryParam(key, serializedParams[key])
      }
    }

    // Query Params
    if (limit !== undefined) {
      requestContext.setQueryParam('limit', ObjectSerializer.serialize(limit, 'number', ''))
    }

    // Query Params
    if (crs !== undefined) {
      requestContext.setQueryParam(
        'crs',
        ObjectSerializer.serialize(
          crs,
          "'http://www.opengis.net/def/crs/OGC/1.3/CRS84' | 'http://www.opengis.net/def/crs/EPSG/0/28992'",
          'uri'
        )
      )
    }

    const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
    if (defaultAuth?.applySecurityAuthentication) {
      await defaultAuth?.applySecurityAuthentication(requestContext)
    }

    return requestContext
  }
}

export class FeaturesApiResponseProcessor {
  /**
   * Unwraps the actual response sent by the server from the response context and deserializes the response content
   * to the expected objects
   *
   * @params response Response returned by the server for a request to search
   * @throws ApiException if the response code was not in [200, 299]
   */
  public async searchWithHttpInfo(response: ResponseContext): Promise<HttpInfo<FeatureCollectionGeoJSON>> {
    const contentType = ObjectSerializer.normalizeMediaType(response.headers['content-type'])
    if (isCodeInRange('200', response.httpStatusCode)) {
      const body: FeatureCollectionGeoJSON = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'FeatureCollectionGeoJSON',
        ''
      ) as FeatureCollectionGeoJSON
      return new HttpInfo(response.httpStatusCode, response.headers, response.body, body)
    }
    if (isCodeInRange('400', response.httpStatusCode)) {
      const body: Exception = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Exception',
        ''
      ) as Exception
      throw new ApiException<Exception>(
        response.httpStatusCode,
        'Bad request: For example, invalid or unknown query parameters.',
        body,
        response.headers
      )
    }
    if (isCodeInRange('404', response.httpStatusCode)) {
      const body: Exception = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Exception',
        ''
      ) as Exception
      throw new ApiException<Exception>(
        response.httpStatusCode,
        'Not found: The requested resource does not exist on the server. For example, a path parameter had an incorrect value.',
        body,
        response.headers
      )
    }
    if (isCodeInRange('406', response.httpStatusCode)) {
      const body: Exception = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Exception',
        ''
      ) as Exception
      throw new ApiException<Exception>(
        response.httpStatusCode,
        'Not acceptable: The requested media type is not supported by this resource.',
        body,
        response.headers
      )
    }
    if (isCodeInRange('500', response.httpStatusCode)) {
      const body: Exception = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Exception',
        ''
      ) as Exception
      throw new ApiException<Exception>(
        response.httpStatusCode,
        'Internal server error: An unexpected server error occurred.',
        body,
        response.headers
      )
    }
    if (isCodeInRange('502', response.httpStatusCode)) {
      const body: Exception = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Exception',
        ''
      ) as Exception
      throw new ApiException<Exception>(
        response.httpStatusCode,
        'Bad Gateway: An unexpected error occurred while forwarding/proxying the request to another server.',
        body,
        response.headers
      )
    }

    // Work around for missing responses in specification, e.g. for petstore.yaml
    if (response.httpStatusCode >= 200 && response.httpStatusCode <= 299) {
      const body: FeatureCollectionGeoJSON = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'FeatureCollectionGeoJSON',
        ''
      ) as FeatureCollectionGeoJSON
      return new HttpInfo(response.httpStatusCode, response.headers, response.body, body)
    }

    throw new ApiException<string | Blob | undefined>(
      response.httpStatusCode,
      'Unknown API Status Code!',
      await response.getBodyAsAny(),
      response.headers
    )
  }
}

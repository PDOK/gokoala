// TODO: better import syntax?
import { BaseAPIRequestFactory, RequiredError, COLLECTION_FORMATS } from './baseapi'
import { Configuration } from '../configuration'
import { RequestContext, HttpMethod, ResponseContext, HttpFile, HttpInfo } from '../http/http'
import { ObjectSerializer } from '../models/ObjectSerializer'
import { ApiException } from './exception'
import { canConsumeForm, isCodeInRange } from '../util'
import { SecurityAuthentication } from '../auth/auth'

import { Collections } from '../models/Collections'
import { Exception } from '../models/Exception'

/**
 * no description
 */
export class CollectionsApiRequestFactory extends BaseAPIRequestFactory {
  /**
   * A list of all collections (geospatial data resources) in this dataset.
   * the collections in the dataset
   * @param f The format of the response. If no value is provided, the standard http rules apply, i.e., the accept header is used to determine the format.  Pre-defined values are \&quot;json\&quot; and \&quot;html\&quot;. The response to other values is determined by the server.
   */
  public async getCollections(f?: 'json' | 'html', _options?: Configuration): Promise<RequestContext> {
    let _config = _options || this.configuration

    // Path Params
    const localVarPath = '/collections'

    // Make Request Context
    const requestContext = _config.baseServer.makeRequestContext(localVarPath, HttpMethod.GET)
    requestContext.setHeaderParam('Accept', 'application/json, */*;q=0.8')

    // Query Params
    if (f !== undefined) {
      requestContext.setQueryParam('f', ObjectSerializer.serialize(f, "'json' | 'html'", ''))
    }

    const defaultAuth: SecurityAuthentication | undefined = _options?.authMethods?.default || this.configuration?.authMethods?.default
    if (defaultAuth?.applySecurityAuthentication) {
      await defaultAuth?.applySecurityAuthentication(requestContext)
    }

    return requestContext
  }
}

export class CollectionsApiResponseProcessor {
  /**
   * Unwraps the actual response sent by the server from the response context and deserializes the response content
   * to the expected objects
   *
   * @params response Response returned by the server for a request to getCollections
   * @throws ApiException if the response code was not in [200, 299]
   */
  public async getCollectionsWithHttpInfo(response: ResponseContext): Promise<HttpInfo<Collections>> {
    const contentType = ObjectSerializer.normalizeMediaType(response.headers['content-type'])
    if (isCodeInRange('200', response.httpStatusCode)) {
      const body: Collections = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Collections',
        ''
      ) as Collections
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
      const body: Collections = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'Collections',
        ''
      ) as Collections
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

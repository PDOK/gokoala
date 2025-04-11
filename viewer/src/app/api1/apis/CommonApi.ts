// TODO: better import syntax?
import { BaseAPIRequestFactory, RequiredError, COLLECTION_FORMATS } from './baseapi'
import { Configuration } from '../configuration'
import { RequestContext, HttpMethod, ResponseContext, HttpFile, HttpInfo } from '../http/http'
import { ObjectSerializer } from '../models/ObjectSerializer'
import { ApiException } from './exception'
import { canConsumeForm, isCodeInRange } from '../util'
import { SecurityAuthentication } from '../auth/auth'

import { ConfClasses } from '../models/ConfClasses'
import { Exception } from '../models/Exception'
import { LandingPage } from '../models/LandingPage'

/**
 * no description
 */
export class CommonApiRequestFactory extends BaseAPIRequestFactory {
  /**
   * A list of all conformance classes specified in a standard that the server conforms to.
   * API conformance definition
   * @param f The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public async getConformanceDeclaration(f?: 'json' | 'html', _options?: Configuration): Promise<RequestContext> {
    let _config = _options || this.configuration

    // Path Params
    const localVarPath = '/conformance'

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

  /**
   * The landing page provides links to the API definition and the conformance statements for this API.
   * Landing page
   * @param f The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public async getLandingPage(f?: 'json' | 'html', _options?: Configuration): Promise<RequestContext> {
    let _config = _options || this.configuration

    // Path Params
    const localVarPath = '/'

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

  /**
   * This document
   * This document
   * @param f The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public async getOpenApi(f?: 'json' | 'html', _options?: Configuration): Promise<RequestContext> {
    let _config = _options || this.configuration

    // Path Params
    const localVarPath = '/api'

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

export class CommonApiResponseProcessor {
  /**
   * Unwraps the actual response sent by the server from the response context and deserializes the response content
   * to the expected objects
   *
   * @params response Response returned by the server for a request to getConformanceDeclaration
   * @throws ApiException if the response code was not in [200, 299]
   */
  public async getConformanceDeclarationWithHttpInfo(response: ResponseContext): Promise<HttpInfo<ConfClasses>> {
    const contentType = ObjectSerializer.normalizeMediaType(response.headers['content-type'])
    if (isCodeInRange('200', response.httpStatusCode)) {
      const body: ConfClasses = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'ConfClasses',
        ''
      ) as ConfClasses
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
      const body: ConfClasses = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'ConfClasses',
        ''
      ) as ConfClasses
      return new HttpInfo(response.httpStatusCode, response.headers, response.body, body)
    }

    throw new ApiException<string | Blob | undefined>(
      response.httpStatusCode,
      'Unknown API Status Code!',
      await response.getBodyAsAny(),
      response.headers
    )
  }

  /**
   * Unwraps the actual response sent by the server from the response context and deserializes the response content
   * to the expected objects
   *
   * @params response Response returned by the server for a request to getLandingPage
   * @throws ApiException if the response code was not in [200, 299]
   */
  public async getLandingPageWithHttpInfo(response: ResponseContext): Promise<HttpInfo<LandingPage>> {
    const contentType = ObjectSerializer.normalizeMediaType(response.headers['content-type'])
    if (isCodeInRange('200', response.httpStatusCode)) {
      const body: LandingPage = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'LandingPage',
        ''
      ) as LandingPage
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
      const body: LandingPage = ObjectSerializer.deserialize(
        ObjectSerializer.parse(await response.body.text(), contentType),
        'LandingPage',
        ''
      ) as LandingPage
      return new HttpInfo(response.httpStatusCode, response.headers, response.body, body)
    }

    throw new ApiException<string | Blob | undefined>(
      response.httpStatusCode,
      'Unknown API Status Code!',
      await response.getBodyAsAny(),
      response.headers
    )
  }

  /**
   * Unwraps the actual response sent by the server from the response context and deserializes the response content
   * to the expected objects
   *
   * @params response Response returned by the server for a request to getOpenApi
   * @throws ApiException if the response code was not in [200, 299]
   */
  public async getOpenApiWithHttpInfo(response: ResponseContext): Promise<HttpInfo<void>> {
    const contentType = ObjectSerializer.normalizeMediaType(response.headers['content-type'])
    if (isCodeInRange('200', response.httpStatusCode)) {
      return new HttpInfo(response.httpStatusCode, response.headers, response.body, undefined)
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
      const body: void = ObjectSerializer.deserialize(ObjectSerializer.parse(await response.body.text(), contentType), 'void', '') as void
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

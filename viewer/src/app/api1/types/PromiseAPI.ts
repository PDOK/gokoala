import { ResponseContext, RequestContext, HttpFile, HttpInfo } from '../http/http'
import { Configuration } from '../configuration'

import { Collection } from '../models/Collection'
import { CollectionLink } from '../models/CollectionLink'
import { Collections } from '../models/Collections'
import { CollectionsCollectionsInner } from '../models/CollectionsCollectionsInner'
import { ConfClasses } from '../models/ConfClasses'
import { Exception } from '../models/Exception'
import { Extent } from '../models/Extent'
import { FeatureCollectionGeoJSON } from '../models/FeatureCollectionGeoJSON'
import { FeatureCollectionJSONFG } from '../models/FeatureCollectionJSONFG'
import { FeatureGeoJSON } from '../models/FeatureGeoJSON'
import { FeatureGeoJSONId } from '../models/FeatureGeoJSONId'
import { FeatureJSONFG } from '../models/FeatureJSONFG'
import { GeometryGeoJSON } from '../models/GeometryGeoJSON'
import { GeometrycollectionGeoJSON } from '../models/GeometrycollectionGeoJSON'
import { IdLink } from '../models/IdLink'
import { Keyword } from '../models/Keyword'
import { LandingPage } from '../models/LandingPage'
import { LinestringGeoJSON } from '../models/LinestringGeoJSON'
import { Link } from '../models/Link'
import { MultilinestringGeoJSON } from '../models/MultilinestringGeoJSON'
import { MultipointGeoJSON } from '../models/MultipointGeoJSON'
import { MultipolygonGeoJSON } from '../models/MultipolygonGeoJSON'
import { PointGeoJSON } from '../models/PointGeoJSON'
import { PolygonGeoJSON } from '../models/PolygonGeoJSON'
import { SearchFunctioneelGebiedParameter } from '../models/SearchFunctioneelGebiedParameter'
import { SearchGeografischGebiedParameter } from '../models/SearchGeografischGebiedParameter'
import { SearchLigplaatsParameter } from '../models/SearchLigplaatsParameter'
import { SearchStandplaatsParameter } from '../models/SearchStandplaatsParameter'
import { SearchVerblijfsobjectParameter } from '../models/SearchVerblijfsobjectParameter'
import { SearchWoonplaatsParameter } from '../models/SearchWoonplaatsParameter'
import { SpatialExtent } from '../models/SpatialExtent'
import { TemporalExtent } from '../models/TemporalExtent'
import { Trs } from '../models/Trs'
import { ObservableCollectionsApi } from './ObservableAPI'

import { CollectionsApiRequestFactory, CollectionsApiResponseProcessor } from '../apis/CollectionsApi'
export class PromiseCollectionsApi {
  private api: ObservableCollectionsApi

  public constructor(
    configuration: Configuration,
    requestFactory?: CollectionsApiRequestFactory,
    responseProcessor?: CollectionsApiResponseProcessor
  ) {
    this.api = new ObservableCollectionsApi(configuration, requestFactory, responseProcessor)
  }

  /**
   * A list of all collections (geospatial data resources) in this dataset.
   * the collections in the dataset
   * @param [f] The format of the response. If no value is provided, the standard http rules apply, i.e., the accept header is used to determine the format.  Pre-defined values are \&quot;json\&quot; and \&quot;html\&quot;. The response to other values is determined by the server.
   */
  public getCollectionsWithHttpInfo(f?: 'json' | 'html', _options?: Configuration): Promise<HttpInfo<Collections>> {
    const result = this.api.getCollectionsWithHttpInfo(f, _options)
    return result.toPromise()
  }

  /**
   * A list of all collections (geospatial data resources) in this dataset.
   * the collections in the dataset
   * @param [f] The format of the response. If no value is provided, the standard http rules apply, i.e., the accept header is used to determine the format.  Pre-defined values are \&quot;json\&quot; and \&quot;html\&quot;. The response to other values is determined by the server.
   */
  public getCollections(f?: 'json' | 'html', _options?: Configuration): Promise<Collections> {
    const result = this.api.getCollections(f, _options)
    return result.toPromise()
  }
}

import { ObservableCommonApi } from './ObservableAPI'

import { CommonApiRequestFactory, CommonApiResponseProcessor } from '../apis/CommonApi'
export class PromiseCommonApi {
  private api: ObservableCommonApi

  public constructor(
    configuration: Configuration,
    requestFactory?: CommonApiRequestFactory,
    responseProcessor?: CommonApiResponseProcessor
  ) {
    this.api = new ObservableCommonApi(configuration, requestFactory, responseProcessor)
  }

  /**
   * A list of all conformance classes specified in a standard that the server conforms to.
   * API conformance definition
   * @param [f] The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public getConformanceDeclarationWithHttpInfo(f?: 'json' | 'html', _options?: Configuration): Promise<HttpInfo<ConfClasses>> {
    const result = this.api.getConformanceDeclarationWithHttpInfo(f, _options)
    return result.toPromise()
  }

  /**
   * A list of all conformance classes specified in a standard that the server conforms to.
   * API conformance definition
   * @param [f] The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public getConformanceDeclaration(f?: 'json' | 'html', _options?: Configuration): Promise<ConfClasses> {
    const result = this.api.getConformanceDeclaration(f, _options)
    return result.toPromise()
  }

  /**
   * The landing page provides links to the API definition and the conformance statements for this API.
   * Landing page
   * @param [f] The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public getLandingPageWithHttpInfo(f?: 'json' | 'html', _options?: Configuration): Promise<HttpInfo<LandingPage>> {
    const result = this.api.getLandingPageWithHttpInfo(f, _options)
    return result.toPromise()
  }

  /**
   * The landing page provides links to the API definition and the conformance statements for this API.
   * Landing page
   * @param [f] The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public getLandingPage(f?: 'json' | 'html', _options?: Configuration): Promise<LandingPage> {
    const result = this.api.getLandingPage(f, _options)
    return result.toPromise()
  }

  /**
   * This document
   * This document
   * @param [f] The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public getOpenApiWithHttpInfo(f?: 'json' | 'html', _options?: Configuration): Promise<HttpInfo<void>> {
    const result = this.api.getOpenApiWithHttpInfo(f, _options)
    return result.toPromise()
  }

  /**
   * This document
   * This document
   * @param [f] The optional f parameter indicates the output format that the server shall provide as part of the response document.  The default format is JSON.
   */
  public getOpenApi(f?: 'json' | 'html', _options?: Configuration): Promise<void> {
    const result = this.api.getOpenApi(f, _options)
    return result.toPromise()
  }
}

import { ObservableFeaturesApi } from './ObservableAPI'

import { FeaturesApiRequestFactory, FeaturesApiResponseProcessor } from '../apis/FeaturesApi'
export class PromiseFeaturesApi {
  private api: ObservableFeaturesApi

  public constructor(
    configuration: Configuration,
    requestFactory?: FeaturesApiRequestFactory,
    responseProcessor?: FeaturesApiResponseProcessor
  ) {
    this.api = new ObservableFeaturesApi(configuration, requestFactory, responseProcessor)
  }

  /**
   * This endpoint allows one to implement autocomplete functionality for location search. The `q` parameter accepts a partial location name and will return all matching locations up to the specified `limit`. The list of search results are offered as features (in GeoJSON, JSON-FG) but contain only minimal information; like a feature ID, highlighted text and a bounding box. When you want to get the full feature you must follow the included link (`href`) in the search result. This allows one to retrieve all properties of the feature and the full geometry from the corresponding OGC API.
   * search features in one or more collections across datasets.
   * @param q The search term(s)
   * @param [functioneelGebied] When provided the functioneel_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the functioneel_gebied collection, for example &#x60;q&#x3D;foo&amp;functioneel_gebied[version]&#x3D;1&amp;functioneel_gebied[relevance]&#x3D;0.5&#x60;
   * @param [geografischGebied] When provided the geografisch_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the geografisch_gebied collection, for example &#x60;q&#x3D;foo&amp;geografisch_gebied[version]&#x3D;1&amp;geografisch_gebied[relevance]&#x3D;0.5&#x60;
   * @param [ligplaats] When provided the ligplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the ligplaats collection, for example &#x60;q&#x3D;foo&amp;ligplaats[version]&#x3D;1&amp;ligplaats[relevance]&#x3D;0.5&#x60;
   * @param [standplaats] When provided the standplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the standplaats collection, for example &#x60;q&#x3D;foo&amp;standplaats[version]&#x3D;1&amp;standplaats[relevance]&#x3D;0.5&#x60;
   * @param [verblijfsobject] When provided the verblijfsobject collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the verblijfsobject collection, for example &#x60;q&#x3D;foo&amp;verblijfsobject[version]&#x3D;1&amp;verblijfsobject[relevance]&#x3D;0.5&#x60;
   * @param [woonplaats] When provided the woonplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the woonplaats collection, for example &#x60;q&#x3D;foo&amp;woonplaats[version]&#x3D;1&amp;woonplaats[relevance]&#x3D;0.5&#x60;
   * @param [limit] The optional limit parameter limits the number of items that are presented in the response document.  Only items are counted that are on the first level of the collection in the response document. Nested objects contained within the explicitly requested items shall not be counted.  Minimum &#x3D; 1. Maximum &#x3D; 50. Default &#x3D; 10.
   * @param [crs] The coordinate reference system of the geometries in the response. Default is WGS84 longitude/latitude
   */
  public searchWithHttpInfo(
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
  ): Promise<HttpInfo<FeatureCollectionGeoJSON>> {
    const result = this.api.searchWithHttpInfo(
      q,
      functioneelGebied,
      geografischGebied,
      ligplaats,
      standplaats,
      verblijfsobject,
      woonplaats,
      limit,
      crs,
      _options
    )
    return result.toPromise()
  }

  /**
   * This endpoint allows one to implement autocomplete functionality for location search. The `q` parameter accepts a partial location name and will return all matching locations up to the specified `limit`. The list of search results are offered as features (in GeoJSON, JSON-FG) but contain only minimal information; like a feature ID, highlighted text and a bounding box. When you want to get the full feature you must follow the included link (`href`) in the search result. This allows one to retrieve all properties of the feature and the full geometry from the corresponding OGC API.
   * search features in one or more collections across datasets.
   * @param q The search term(s)
   * @param [functioneelGebied] When provided the functioneel_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the functioneel_gebied collection, for example &#x60;q&#x3D;foo&amp;functioneel_gebied[version]&#x3D;1&amp;functioneel_gebied[relevance]&#x3D;0.5&#x60;
   * @param [geografischGebied] When provided the geografisch_gebied collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the geografisch_gebied collection, for example &#x60;q&#x3D;foo&amp;geografisch_gebied[version]&#x3D;1&amp;geografisch_gebied[relevance]&#x3D;0.5&#x60;
   * @param [ligplaats] When provided the ligplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the ligplaats collection, for example &#x60;q&#x3D;foo&amp;ligplaats[version]&#x3D;1&amp;ligplaats[relevance]&#x3D;0.5&#x60;
   * @param [standplaats] When provided the standplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the standplaats collection, for example &#x60;q&#x3D;foo&amp;standplaats[version]&#x3D;1&amp;standplaats[relevance]&#x3D;0.5&#x60;
   * @param [verblijfsobject] When provided the verblijfsobject collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the verblijfsobject collection, for example &#x60;q&#x3D;foo&amp;verblijfsobject[version]&#x3D;1&amp;verblijfsobject[relevance]&#x3D;0.5&#x60;
   * @param [woonplaats] When provided the woonplaats collection is included in the search. This parameter should be provided as a [deep object](https://swagger.io/docs/specification/v3_0/serialization/#query-parameters) containing the version and relevance of the woonplaats collection, for example &#x60;q&#x3D;foo&amp;woonplaats[version]&#x3D;1&amp;woonplaats[relevance]&#x3D;0.5&#x60;
   * @param [limit] The optional limit parameter limits the number of items that are presented in the response document.  Only items are counted that are on the first level of the collection in the response document. Nested objects contained within the explicitly requested items shall not be counted.  Minimum &#x3D; 1. Maximum &#x3D; 50. Default &#x3D; 10.
   * @param [crs] The coordinate reference system of the geometries in the response. Default is WGS84 longitude/latitude
   */
  public search(
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
  ): Promise<FeatureCollectionGeoJSON> {
    const result = this.api.search(
      q,
      functioneelGebied,
      geografischGebied,
      ligplaats,
      standplaats,
      verblijfsobject,
      woonplaats,
      limit,
      crs,
      _options
    )
    return result.toPromise()
  }
}

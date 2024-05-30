# Viewer

This viewer is available as a [WebComponent](https://developer.mozilla.org/en-US/docs/Web/API/Web_components) and amongst others able to:

- display vector tiles, with object information
- display GeoJSON features on a map
- render a legend of a Mapbox style

See [demo with samples](https://pdok.github.io/gokoala/).

## Embedding a vectortile map

Build the project and copy the style and javascript files into the application embedding the download component

Embed the webcomponent 'app-vectortile-view' in your web application

- load styles and javascript in your html

  ```html
  <link rel="stylesheet" type="text/css" href="view-component/styles.css" />
  <script type="text/javascript" src="view-component/main.js"></script>
  <script type="text/javascript" src="view-component/polyfills.js"></script>
  <script type="text/javascript" src="view-component/runtime.js"></script>

  <app-vectortile-view
    style="width: 800px; height: 600px;"
    tile-url="https://api.pdok.nl/lv/bag/ogc/v0_1/tiles/NetherlandsRDNewQuad"
    zoom="12"
    center-x="5.3896944"
    center-y="52.1562499">
  </app-vectortile-view>
  ```

## Parameters for vectortile view

The vectortile view comnponent has the follwing parameters:

- **tileUrl** : Url to OGC vector tile service
- **styleUrl** Url to vector Mapbox tile style.
- **id** id for map
- **zoom** initial zoom level

The following values are emitted:

- **currentZoomLevel**
- **activeFeature**
- **activeTileUrl**
- **centerX**
- **centerY**

## Embedding a OGC API feature view

<app-feature-view
      id="featuresample"
      mode="auto"
      fill-color="rgba(0,0,255,0)"
      items-url="https://api.pdok.nl/lv/bgt/ogc/v1/collections/pand/items/1">
</app-feature-view>

## Feature view parameters

The view comnponent has the follwing parameters

- **itemsUrl**: A OGC API url as dataset for the features to show
- **backgroundMap**: Openstreetmap is used as default backgroundmap. Use value "BRT"to use Dutch "brt achtergrondkaart" as background
- **fillColor**: fill color (hex or RBG) for the features If not specified is used 'rgba(0,0,255)' use e.g. "rgba(0,0,255,0)" for a transparent fill
- **strokeColor**: Stroke color of the feature default color is '#3399CC'
- **mode**: Operation mode is 'default' or 'auto'. If 'auto' is used the bounding box of the view is emitted as boundingbox, and no buttons are visible.
- **showBoundingBoxButton**: in default mode the boundingbox select button is showed, hide 'show-bounding-box-button' is needed
- **showFillExtentButton**: in default mode the button to fill the view with features is not showed. Activate 'show-fill-extent-button' is needed
- **projection**: projection in opengis style e.g. '<http://www.opengis.net/def/crs/EPSG/0/4258>'
- **labelField**: field is show as label and feature is clickable. if not specified a popup is shown when hovering over feature

The following values are emitted:

- **box**
- **activeFeature**

```html
<link rel="stylesheet" type="text/css" href="view-component/styles.css" />
<script type="text/javascript" src="view-component/main.js"></script>
<script type="text/javascript" src="view-component/polyfills.js"></script>
<script type="text/javascript" src="view-component/runtime.js"></script>

<app-legend-view
  id="legendadminunit"
  style-url="https://api.pdok.nl/kadaster/bestuurlijkegebieden/ogc/v1_0-preprod/styles/bestuurlijkegebieden_standaardvisualisatie?f=json">
</app-legend-view>
```

# Legend Parameters

The legend has the following parameters:

- **style-url**: This is the URL to the Mapbox style that serves as input for the legend.

- **title-items**: By default, the source layer names are used to name legend items. However, this parameter can be used to split legend items based on different attributes.

Default layers are used for legend items. Attributes can be specified to create distinct items. For example, for the Dutch BGT, `titleItems = "type,plus_type,functie,fysiek_voorkomen,openbareruimtetype"` can be used. When `titleItems = "id"` is used, the "id" for the layer (layer name) is used to name the legend items.

Legend uses is shown in the [example directory](../examples/)

## Development server

Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The application will automatically reload if you change any of the source files.

## Build

Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.

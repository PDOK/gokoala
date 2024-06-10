# Viewer

This viewer is available as a [WebComponent](https://developer.mozilla.org/en-US/docs/Web/API/Web_components) and amongst others able to:

- display vector tiles, with object information
- display GeoJSON features on a map
- render a legend of a Mapbox style

See [demo with samples](https://pdok.github.io/gokoala/).

## Parameters for vectortile view

The `<app-vectortile-view>` component has the following parameters:

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

## Feature view parameters

The `<app-feature-view>` component has the following parameters

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

## Legend view parameters

The `<app-legend-view>` component has the following parameters:

- **style-url**: This is the URL to the Mapbox style that serves as input for the legend.
- **title-items**: By default, the source layer names are used to name legend items. However, this parameter can be used to split legend items based on different attributes.

Default layers are used for legend items. Attributes can be specified to create distinct items. For example, for the Dutch BGT, `titleItems = "type,plus_type,functie,fysiek_voorkomen,openbareruimtetype"` can be used. When `titleItems = "id"` is used, the "id" for the layer (layer name) is used to name the legend items.
Legend uses is shown in the [example directory](./examples)

## Development server

Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The application will automatically reload if you change any of the source files.

## Build

Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.

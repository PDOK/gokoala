# Vectortile map and legend component

See [demo with samples](https://pdok.github.io/gokoala/)

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

## Embedding a vectortile legend

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

see [index.html](./src/index.html) for other samples used in the [demo](https://pdok.github.io/gokoala/)

## Development server

Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The application will automatically reload if you change any of the source files.

## Build

Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.
